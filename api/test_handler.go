package api

import (
	"github.com/dapr-platform/common"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func InitTestRoute(r chi.Router) {
	r.Post(common.BASE_CONTEXT+"/sim-send/{pubsub}/{topic}", pubsubTopicSendHandler)

}

// @Summary 模拟发送到Topic消息
// @Description 模拟发送到前端消息
// @Tags MessageSim
// @Param pubsub path string true "pubsub"
// @Param topic path string true "web"
// @Param msg  body map[string]interface{} true "模拟消息"
// @Produce  json
// @Success 20000 {object} common.Response "ok"
// @Failure 50000 {object} common.Response "错误code和错误信息"
// @Router /sim-send/{pubsub}/{topic} [post]
func pubsubTopicSendHandler(w http.ResponseWriter, r *http.Request) {
	pubsub := chi.URLParam(r, "pubsub")
	topic := chi.URLParam(r, "topic")
	if pubsub == "" || topic == "" {
		common.HttpResult(w, common.ErrParam.AppendMsg("path variable is blank"))
		return
	}
	var msg map[string]interface{}
	err := common.ReadRequestBody(r, &msg)
	if err != nil {
		common.HttpResult(w, common.ErrParam.AppendMsg("read body error").AppendMsg(err.Error()))
		return
	}
	err = common.GetDaprClient().PublishEvent(r.Context(), pubsub, topic, msg)
	if err != nil {
		common.HttpResult(w, common.ErrParam.AppendMsg("PublishEvent error").AppendMsg(err.Error()))
		return
	}
	common.HttpSuccess(w, common.OK)
}
