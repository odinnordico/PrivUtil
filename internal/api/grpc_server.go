package api

import (
	pb "github.com/odinnordico/privutil/proto"
)

type Server struct {
	pb.UnimplementedPrivUtilServiceServer
}

func NewServer() *Server {
	return &Server{}
}
