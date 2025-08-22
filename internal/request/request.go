package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpFromScratch/internal/headers"
	"io"
	"strconv"
)

// what we are doing in this file of code

// When we read,
// all we're doing is moving the data from the reader (which in the case of HTTP is a network connection,
// but it could be a file as well, our code is agnostic) into our program. When we parse,
// we're taking that data and interpreting it (moving it from a []byte to a
// RequestLine struct). Once its parsed, we can discard it from the buffer to save memory.

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState string
const(
	StateInit parserState = "init"
	StateDone parserState = "done"
	StateError parserState = "Error"
	StateBody parserState = "Body"
	StateHeaders parserState = "headers"
)

type Request struct {
	RequestLine RequestLine
	Headers headers.Headers
	state parserState
	Body string
}

func GetInt(headers *headers.Headers,name string,DefaultValue int)(int){
	valueStr,ok := headers.Get(name)
	if !ok{
		return DefaultValue
	}
	value,err := strconv.Atoi(valueStr)
	if err!=nil{
		return DefaultValue
	}
	return value
}

func newRequest() *Request{
	return &Request{
		state: StateInit,
		Headers: *headers.NewHeaders(),
		Body: "",
	}
}

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported HTTP version")
var ERROR_REQUEST_IN_ERROR_STATE = fmt.Errorf("Request in error state")
var SEPARATOR =[]byte("\r\n")

// func (r *RequestLine) ValidHTTP() bool{
// 	return r.HttpVersion == "HTTP/1.1"
// }

func parseRequestLine(b []byte)(*RequestLine,int,error){
	// fmt.Printf("hello======================================")
	idx := bytes.Index(b,SEPARATOR)
	if( idx==-1){
		fmt.Printf("idx = -1")
		return nil,0,nil
	}
	startLine := b[:idx]
	// fmt.Printf("THis is start line------------------%s",string(startLine))
	read := idx+len(SEPARATOR)
	parts := bytes.Split(startLine,[]byte(" "))
	if len(parts)!=3{
		return nil,0,ERROR_MALFORMED_REQUEST_LINE
	}
	httpParts := bytes.Split(parts[2],[]byte("/"))
	if len(httpParts) != 2 || string(httpParts[0])!="HTTP" || string(httpParts[1])!="1.1"{
		return nil,0,ERROR_UNSUPPORTED_HTTP_VERSION
	}
	rl := &RequestLine{
		Method: string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion: string(httpParts[1]),
	}
	return rl,read,nil
}
func (r *Request) hasBody()bool{
	length := GetInt(&r.Headers,"content-length",0)
	return length > 0
}
func (r *Request) parse(data []byte) (int,error){
	read:=0
outer:
	for{
		currentData := data[read:]
		if len(currentData) == 0{
			break outer
		}
		switch r.state{
		case StateError:
			return 0,ERROR_REQUEST_IN_ERROR_STATE
		case StateInit:
			// we need to parse header
			rl,n,err := parseRequestLine(currentData)
			if err!=nil{
				r.state = StateError
				return 0,err
			}
			if n ==0{
				break outer
			} 
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders
			
		case StateHeaders:
			n,done,err := r.Headers.Parse(currentData)
			if err != nil{
				return 0,err
			}
			if n==0{
				break outer
			}
			read += n
			if done{
				if r.hasBody(){
					r.state = StateBody
				}else{
					r.state = StateDone
				}
			}
		case StateBody:
			length := GetInt(&r.Headers,"content-length",0)
			if length == 0{
				r.state = StateDone
				break
			}
			remaining := min(length - len(r.Body),len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining
			if len(r.Body) == length{
				r.state = StateDone
			}
		case StateDone:
			break outer
		default:
			panic("Somehow we have programmed poorly")
		}
	}
	return read,nil
}
func (r *Request) done() bool{
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader)(*Request,error){
	request := newRequest()
	buf := make([]byte,1024)
	bufLen := 0
	for !request.done(){
		// read from network stream
		n,err := reader.Read(buf[bufLen:])
		// fmt.Printf("DEBUG raw line: %q\n", string(buf[:n]))
		if err!=nil{
			if errors.Is(err,io.EOF){
				request.state=StateDone
				break
			}
			return nil,err
		}
		// add not of bytes Read to bufLen
		bufLen += n

		readN,err := request.parse(buf[:bufLen])
		// after parsing some of the data in buffer
		// change buffer to un-parsed data

		// readN of buffer gets parsed, so the remaining readN to bufLen remains unparsed
		// hence update buf with it.
		copy(buf,buf[readN:bufLen])
		bufLen -= readN
	}
	return request,nil;
}