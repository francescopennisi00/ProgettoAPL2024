package communication_grpc

import wmsUm "wms_microservice/proto"

type UmWmsServer struct {
	wmsUm.UnimplementedWMSUmServer
}
