package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"

	"github.com/superwhys/billiard-helper/models/errcode"
	"github.com/superwhys/billiard-helper/models/request"
	"github.com/superwhys/billiard-helper/models/response"
	"github.com/superwhys/billiard-helper/ports"
)

func RoomGroupRouter(roomSvc ports.RoomService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/rooms"),
		ginutils.WithHandler(http.MethodPost, "/create", RoomCreateHandler(roomSvc)),
		ginutils.WithHandler(http.MethodGet, "/:room_id", RoomDetailHandler(roomSvc)),
		ginutils.WithHandler(http.MethodGet, "/list", RoomListHandler(roomSvc)),
		ginutils.WithHandler(http.MethodDelete, "/:room_id", RoomDeleteHandler(roomSvc)),
	)
}

// RoomCreateHandler 处理创建房间
// @Summary 创建房间
// @Description 创建房间
// @Tags Room
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateRoomRequest true "创建房间请求体"
// @Success 200 {object} ginutils.Ret[response.Room]
// @Router /rooms/create [post]
func RoomCreateHandler(roomSvc ports.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *request.CreateRoomRequest) {
		room, err := roomSvc.CreateRoom(c.Request.Context(), req)
		if handleRouterError(c, err, "room create handler error", errcode.ErrCodeCreateRoomFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseWithData(room))
	})
}

// RoomDetailHandler 处理房间详情
// @Summary 获取房间详情
// @Description 获取房间详情
// @Tags Room
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.GetRoomRequest true "获取房间详情请求体"
// @Success 200 {object} ginutils.Ret[response.Room]
// @Router /rooms/:room_id [get]
func RoomDetailHandler(roomSvc ports.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *request.GetRoomRequest) {
		room, err := roomSvc.GetRoom(c.Request.Context(), req)
		if handleRouterError(c, err, "room detail handler error", errcode.ErrCodeGetRoomFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseWithData(room))
	})
}

// RoomListHandler 处理用户房间列表
// @Summary 获取用户房间列表
// @Description 获取用户房间列表
// @Tags Room
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.GetUserRoomsRequest true "获取用户房间列表请求体"
// @Success 200 {object} ginutils.Ret[[]response.Room]
// @Router /rooms/list [get]
func RoomListHandler(roomSvc ports.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *request.GetUserRoomsRequest) {
		rooms, err := roomSvc.GetUserRooms(c.Request.Context(), req)
		if handleRouterError(c, err, "room list handler error", errcode.ErrCodeGetUserRoomsFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseWithData(rooms))
	})
}

// RoomDeleteHandler 处理删除房间
// @Summary 删除房间
// @Description 删除房间
// @Tags Room
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.DeleteRoomRequest true "删除房间请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /rooms/:room_id [delete]
func RoomDeleteHandler(roomSvc ports.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *request.DeleteRoomRequest) {
		err := roomSvc.DeleteRoom(c.Request.Context(), req)
		if handleRouterError(c, err, "room delete handler error", errcode.ErrCodeDeleteRoomFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}
