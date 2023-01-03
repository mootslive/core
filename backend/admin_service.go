package backend

import (
	"context"
	"log"

	"github.com/bufbuild/connect-go"
	mootslivepbv1 "github.com/mootslive/mono/proto/mootslive/v1"
	"github.com/mootslive/mono/proto/mootslive/v1/mootslivepbv1connect"
)

type AdminService struct {
	mootslivepbv1connect.UnimplementedAdminServiceHandler
}

func (as *AdminService) GetStatus(
	ctx context.Context,
	req *connect.Request[mootslivepbv1.GetStatusRequest],
) (*connect.Response[mootslivepbv1.GetStatusResponse], error) {
	log.Println("Request headers: ", req.Header())
	res := connect.NewResponse(&mootslivepbv1.GetStatusResponse{
		XClacksOverhead: "GNU Corey Kendall",
	})
	return res, nil
}
