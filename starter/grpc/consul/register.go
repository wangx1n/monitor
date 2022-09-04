package consul

import consulServerimport "monitor/handler/consulServer"

func RegisterToConsul() {
	consulServerimport.RegitserService("192.168.31.72:8500", &consulServerimport.ConsulService{
		Name: "helloworld",
		Tag:  []string{"helloworld"},
		IP:   "192.168.31.88",
		Port: 50051,
	})
}
