package server

import (
	"golang.org/x/sys/unix"
	"log"
	"net"
	"sync"
	"bytes"
	"net/http"
	"bufio"
	"os"
	"hlcup_epoll/handlers"
	"regexp"
	"fmt"
	"hlcup_epoll/services"
	"time"
	"runtime"
)

var getUserRegexp *regexp.Regexp
var getLocationRegexp *regexp.Regexp
var getVisitRegexp *regexp.Regexp
var getVisitedPlacesRegexp *regexp.Regexp
var getPlaceAvgMarkRegexp *regexp.Regexp
var createUserRegexp *regexp.Regexp
var createLocationRegexp *regexp.Regexp
var createVisitRegexp *regexp.Regexp
var idRegexp *regexp.Regexp

var responseBytes []byte
var responseCode int

type Server struct {
	errorLogger        *log.Logger
	infoLogger        *log.Logger
	port               int
	userApiHandler     *handlers.UserApiHandler
	locationApiHandler *handlers.LocationApiHandler
	visitApiHandler    *handlers.VisitApiHandler
}

func NewServer(port int, dataPath string, optionsPath string) *Server {

	errorLogger := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Llongfile)
	infoLogger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)

	startTime := time.Now()

	compileRegexp()

	storage := services.NewStorage(errorLogger, infoLogger)

	waitGroup := new(sync.WaitGroup)

	storage.Init(dataPath, 4, waitGroup)

	waitGroup.Wait()

	infoLogger.Println(fmt.Sprintf("Storage has been filled. Duration %s", time.Since(startTime).String()))

	printMemUsage()
	runtime.GC()
	printMemUsage()

	server := new(Server)

	server.errorLogger = errorLogger
	server.infoLogger = infoLogger
	server.port = port
	server.userApiHandler = handlers.NewUserApiHandler(storage, errorLogger, infoLogger, optionsPath)
	server.locationApiHandler = handlers.NewLocationApiHandler(storage, errorLogger, infoLogger, optionsPath)
	server.visitApiHandler = handlers.NewVisitApiHandler(storage, errorLogger, infoLogger)

	return server
}

func compileRegexp() {

	getUserRegexp = regexp.MustCompile("^/users/\\d+.*")
	getLocationRegexp = regexp.MustCompile("^/locations/\\d+.*")
	getVisitRegexp = regexp.MustCompile("^/visits/\\d+.*")

	getVisitedPlacesRegexp = regexp.MustCompile("^/users/\\d+/visits.*")
	getPlaceAvgMarkRegexp = regexp.MustCompile("^/locations/\\d+/avg.*")

	createUserRegexp = regexp.MustCompile("^/users/new.*")
	createLocationRegexp = regexp.MustCompile("^/locations/new.*")
	createVisitRegexp = regexp.MustCompile("^/visits/new.*")

	idRegexp = regexp.MustCompile("\\d+")
}

func (server *Server) Run() {

	cpuCount := runtime.NumCPU()

	waitGroup := new(sync.WaitGroup)

	for i := 0; i < cpuCount; i++ {

		waitGroup.Add(1)

		connectionEpollFd := server.handleConnection()

		server.handleAccept(connectionEpollFd)
	}

	server.infoLogger.Println(fmt.Sprintf("Server is listening on %d CPUs", cpuCount))

	waitGroup.Wait()
}

func (server *Server) handleConnection() int {

	connectionEpollFd, err := unix.EpollCreate1(0)

	if err != nil {
		server.errorLogger.Fatalln(err)
	}

	events := make([]unix.EpollEvent, 1024)

	go func() {

		for {

			countEvents, err := unix.EpollWait(connectionEpollFd, events, -1)

			if err != nil {
				server.errorLogger.Fatalln(err)
			}

			for eventIndex := 0; eventIndex < countEvents; eventIndex++ {

				buffer := make([]byte, 1024)
				event := events[eventIndex]

				for {

					countBytes, err := unix.Read(int(event.Fd), buffer)

					if countBytes == -1 && (err == unix.EAGAIN || err == unix.EWOULDBLOCK) {
						break

					} else if countBytes == -1 {
						unix.Close(int(events[eventIndex].Fd))
						server.errorLogger.Fatalln(err)
					}
				}

				requestReader := bytes.NewReader(buffer)

				bufReader := bufio.NewReader(requestReader)

				httpRequest, err := http.ReadRequest(bufReader)

				if err != nil {
					server.errorLogger.Fatalln(err)
				}

				switch {
				case getVisitedPlacesRegexp.MatchString(httpRequest.RequestURI) && httpRequest.Method == http.MethodGet:
					responseBytes, responseCode = server.userApiHandler.GetVisitedPlaces(httpRequest, idRegexp.FindString(httpRequest.RequestURI))
					break
				case getPlaceAvgMarkRegexp.MatchString(httpRequest.RequestURI) && httpRequest.Method == http.MethodGet:
					responseBytes, responseCode = server.locationApiHandler.GetAverageMark(httpRequest, idRegexp.FindString(httpRequest.RequestURI))
					break
				case getUserRegexp.MatchString(httpRequest.RequestURI) && httpRequest.Method == http.MethodGet:
					responseBytes, responseCode = server.userApiHandler.GetById(httpRequest, idRegexp.FindString(httpRequest.RequestURI))
					break
				case getLocationRegexp.MatchString(httpRequest.RequestURI) && httpRequest.Method == http.MethodGet:
					responseBytes, responseCode = server.locationApiHandler.GetById(httpRequest, idRegexp.FindString(httpRequest.RequestURI))
					break
				case getVisitRegexp.MatchString(httpRequest.RequestURI) && httpRequest.Method == http.MethodGet:
					responseBytes, responseCode = server.visitApiHandler.GetById(httpRequest, idRegexp.FindString(httpRequest.RequestURI))
					break
				case createUserRegexp.MatchString(httpRequest.RequestURI) && httpRequest.Method == http.MethodPost:
					responseBytes, responseCode = server.userApiHandler.Create(httpRequest)
					break
				case createLocationRegexp.MatchString(httpRequest.RequestURI) && httpRequest.Method == http.MethodPost:
					responseBytes, responseCode = server.locationApiHandler.Create(httpRequest)
					break
				case createVisitRegexp.MatchString(httpRequest.RequestURI) && httpRequest.Method == http.MethodPost:
					responseBytes, responseCode = server.visitApiHandler.Create(httpRequest)
					break
				case getUserRegexp.MatchString(httpRequest.RequestURI) && httpRequest.Method == http.MethodPost:
					responseBytes, responseCode = server.userApiHandler.Update(httpRequest, idRegexp.FindString(httpRequest.RequestURI))
					break
				case getLocationRegexp.MatchString(httpRequest.RequestURI) && httpRequest.Method == http.MethodPost:
					responseBytes, responseCode = server.locationApiHandler.Update(httpRequest, idRegexp.FindString(httpRequest.RequestURI))
					break
				case getVisitRegexp.MatchString(httpRequest.RequestURI) && httpRequest.Method == http.MethodPost:
					responseBytes, responseCode = server.visitApiHandler.Update(httpRequest, idRegexp.FindString(httpRequest.RequestURI))
					break

				default:
					responseCode = 404
				}

				if responseCode == 404 {
					err = server.returnNotFound(event.Fd)

				} else if responseCode == 400 {
					err = server.returnBadRequest(event.Fd)

				} else {
					err = server.returnOk(event.Fd, responseBytes)
				}

				if err != nil {
					unix.Close(int(events[eventIndex].Fd))
					server.errorLogger.Fatalln(err)
				}

				err = unix.Close(int(events[eventIndex].Fd))

				if err != nil {
					server.errorLogger.Fatalln(err)
				}
			}
		}

	}()

	return connectionEpollFd
}

func (server *Server) returnNotFound(connectionFd int32) error {

	response := make([]byte, 0)

	response = append(response, fmt.Sprintf("HTTP/1.1 404 OK\r\nContent-Type: text/plain\r\nConnection: close\r\nContent-Length: %d\r\n\r\nNot Found", len("Not Found"))...)

	_, err := unix.Write(int(connectionFd), response)

	return err
}

func (server *Server) returnBadRequest(connectionFd int32) error {

	response := make([]byte, 0)

	response = append(response, fmt.Sprintf("HTTP/1.1 400 OK\r\nContent-Type: text/plain\r\nConnection: close\r\nContent-Length: %d\r\n\r\nBad Request", len("Bad Request"))...)

	_, err := unix.Write(int(connectionFd), response)

	return err
}

func (server *Server) returnOk(connectionFd int32, data []byte) error {

	response := make([]byte, 0)

	response = append(response, fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nConnection: close\r\nContent-Length: %d\r\n\r\n%s", len(data), string(data))...)

	_, err := unix.Write(int(connectionFd), response)

	return err
}

func (server *Server) handleAccept(connectionEpollFd int) {

	socketAddr := &unix.SockaddrInet4{Port: server.port}
	copy(socketAddr.Addr[:], net.ParseIP(`0.0.0.0`).To4())

	socketFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK, 0)

	if err != nil {
		server.errorLogger.Fatalln(err)
	}

	err = unix.SetNonblock(socketFd, true)

	if err != nil {
		unix.Close(socketFd)
		server.errorLogger.Fatalln(err)
	}

	err = unix.SetsockoptInt(socketFd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)

	if err != nil {
		unix.Close(socketFd)
		server.errorLogger.Fatalln(err)
	}

	err = unix.SetsockoptInt(socketFd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)

	if err != nil {
		unix.Close(socketFd)
		server.errorLogger.Fatalln(err)
	}

	err = unix.SetsockoptInt(socketFd, unix.SOL_TCP, unix.TCP_QUICKACK, 1)

	if err != nil {
		unix.Close(socketFd)
		server.errorLogger.Fatalln(err)
	}

	err = unix.SetsockoptInt(socketFd, unix.SOL_TCP, unix.TCP_NODELAY, 1)

	if err != nil {
		unix.Close(socketFd)
		server.errorLogger.Fatalln(err)
	}

	err = unix.Bind(socketFd, socketAddr)

	if err != nil {
		unix.Close(socketFd)
		server.errorLogger.Fatalln(err)
	}

	err = unix.Listen(socketFd, unix.SOMAXCONN)

	if err != nil {
		unix.Close(socketFd)
		server.errorLogger.Fatalln(err)
	}

	socketEpollFd, err := unix.EpollCreate1(0)

	if err != nil {
		unix.Close(socketFd)
		server.errorLogger.Fatalln(err)
	}

	epollEvent := &unix.EpollEvent{Events: unix.EPOLLIN | unix.EPOLLEXCLUSIVE | unix.EPOLLET, Fd: int32(socketFd)}

	err = unix.EpollCtl(socketEpollFd, unix.EPOLL_CTL_ADD, socketFd, epollEvent)

	if err != nil {
		unix.Close(socketFd)
		server.errorLogger.Fatalln(err)
	}

	events := make([]unix.EpollEvent, 1024)

	go func() {

		for {

			countEvents, err := unix.EpollWait(socketEpollFd, events, -1)

			if err != nil {
				unix.Close(socketFd)
				unix.Close(socketEpollFd)
				server.errorLogger.Fatalln(err)
			}

			for eventIndex := 0; eventIndex < countEvents; eventIndex++ {

				for {

					connectionFd, _, err := unix.Accept(socketFd)

					if connectionFd == -1 && err == unix.EAGAIN {
						break

					} else if connectionFd == -1 {
						unix.Close(socketFd)
						unix.Close(socketEpollFd)
						unix.Close(connectionFd)
						server.errorLogger.Fatalln(err)
					}

					if err != nil {
						unix.Close(socketFd)
						unix.Close(socketEpollFd)
						server.errorLogger.Fatalln(err)
					}

					err = unix.SetNonblock(connectionFd, true)

					if err != nil {
						unix.Close(socketFd)
						unix.Close(socketEpollFd)
						unix.Close(connectionFd)
						server.errorLogger.Fatalln(err)
					}

					epollConnectionEvent := &unix.EpollEvent{Events: unix.EPOLLET | unix.EPOLLIN | unix.EPOLLONESHOT, Fd: int32(connectionFd)}

					err = unix.EpollCtl(connectionEpollFd, unix.EPOLL_CTL_ADD, connectionFd, epollConnectionEvent)

					if err != nil {
						unix.Close(socketFd)
						unix.Close(socketEpollFd)
						unix.Close(connectionFd)
						server.errorLogger.Fatalln(err)
					}
				}
			}
		}

	}()
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}