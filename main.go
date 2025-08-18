package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func getLinesChannel(f io.ReadCloser)<-chan string{
		
}

func main(){
	fmt.Println("I hope i get that job!")
	file,err := os.Open("./message.txt")
	if(err!=nil){
		log.Fatal("Error: ",err)
	}
	// read the text in message.txt 8 byte at a time
	str := ""
	for{
		data := make([]byte,8)
		// n represents the number of bytes read could be less than 8 or 8
		n,err := file.Read(data)
		if err!=nil{
			break
		}
		strtemp := ""
		for i := 0;i<n;i++{
			if(string(data[i]) != "\n"){
				strtemp += string(data[i])
			}
			if(string(data[i]) == "\n"){
				continue;
			}
		}
		str += strtemp
	}
	fmt.Printf("file: %s\n",str)


}