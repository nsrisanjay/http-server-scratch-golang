package headers

import (
	"bytes"
	"fmt"
	"strings"
)

func isToken(str []byte) bool{
	for _, ch := range str{
		found := false
		if ch >='A' && ch<='Z' || ch>='a' && ch<='z'|| ch >='0' && ch<='9'{
			found = true
	   }
		switch ch{
		case '!','#','$','%','&','\'','*','+','-','.','^','_','`','|','~':
			found =  true
		}
		if !found{
			return false
		}
	}
	return true
}

var rn = []byte("\r\n")
func parseHeader(fieldLine []byte)(string,string,error){
	parts := bytes.SplitN(fieldLine,[]byte(":"),2)
	if len(parts)!=2{
		return "","",fmt.Errorf("malformed error")
	}
	name := parts[0]
	value := bytes.TrimSpace(parts[1])
	if bytes.HasSuffix(name,[]byte(" ")){
		// shouldnt have space after key, but has space
		return "","",fmt.Errorf("malformed field name")
	}
	return string(name),string(value),nil
}

type Headers struct{
	headers map[string]string
}
func (h *Headers) Get(name string) string{
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name string,value string){
	name = strings.ToLower(name)
	v,ok := h.headers[name]
	if ok{
		h.headers[name] = fmt.Sprintf("%s,%s",v,value)
	}else{
		h.headers[name] = value
	}
}

func NewHeaders() *Headers{
	return &Headers{
		headers: map[string]string{},
	}
}



func (h Headers) Parse(data []byte) (int,bool,error){
	read:=0
	done := false
	for{
		idx := bytes.Index(data[read:],rn)
		if idx == -1{
			// the header doesnt end!!
			break
		}
		// hit the empty header
		if idx == 0{
			done = true
			read += len(rn)
			break
		}
		name,value,err := parseHeader(data[read:read+idx])
		if err!=nil{
			return 0, false,err
		}
		if isToken([]byte(name)) == false{
			return 0,false,fmt.Errorf("malformed header name(key)")
		}
		read += idx+len(rn)
		h.headers[name] = value
	}
	return read,done,nil
}