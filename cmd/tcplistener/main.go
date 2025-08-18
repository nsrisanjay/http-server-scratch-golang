package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser)<-chan string{
	// make a channel of strings
	ch := make(chan string,1)
	go func() {
		defer close(ch)
		defer f.Close()
		// read the text in message.txt 8 byte at a time
		str := ""
		for{
			data := make([]byte,8)
			n,err := f.Read(data)
			if err!=nil{
				break
			}
			// n represents the number of bytes read could be less than 8 or 8
			data = data[:n]
			// return the index at which '\n' lies in the byte of data
			index := bytes.IndexByte(data,'\n')
			if index!=-1 {
				str += string(data[:index])
				// now save data after the '\n'
				data = data[index+1:]
				// fmt.Printf("read : %s\n",str)
				ch<-str
				str = ""
			}
			// append the saved data
			str += string(data)
		}
		if len(str)!=0{
			// fmt.Printf("file: %s\n",str)
			ch<-str
		}
	}()
	return ch
}

func main(){
	fmt.Println("I hope i get that job!")
	// create a tcp listener on port 42069
	listener,err := net.Listen("tcp",":42069")
	if err!=nil{
		log.Fatal("Error",err)
	}
	for{
		conn,err := listener.Accept()
		if err!=nil{
			log.Fatal("error: ",err)
		}
		fmt.Println("TCP connection has been accepted")
		lines := getLinesChannel(conn)
		for line := range lines{
			fmt.Printf("%s\n",line)
		}
		fmt.Println("TCP connection has been closed")
	}
}