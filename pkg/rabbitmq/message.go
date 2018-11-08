package rabbitmq

import (
	"time"
)

const MsgTypePing = "ping"
const MsgTypeKubeApiRequest = "kube_api_request"
const MsgTypeStreamResourceLogs = "stream_resource_logs"
const MsgTypeTaskKubeDeploy = "task_kube_deploy"
const MsgTypeTaskKubeRunJob = "task_kube_run_job"
const MsgTypeTaskGet = "task_get"
const MsgTypeTaskStreamLogs = "task_stream_logs"

type KubeApiRequest struct {
	Method string `json:"method"`
	URI    string `json:"uri"`
	Body   string `json:"body"`
}

type KubeApiResponse struct {
	HttpCode    string `json:"http_code"`
	HttpMessage string `json:"http_message"`
	Body        string `json:"body"`
}

type StreamResourceLogs struct {
	StreamId  string    `json:"stream_id"`
	Since     time.Time `json:"since"`
	Duration  uint      `json:"duration"`
	Period    uint      `json:"period"`
	Resource  string    `json:"resource"`
	Namespace string    `json:"namespace "`
	Name      string    `json:"name"`
}

type TaskKubeDeploy struct {
	TaskId   string `json:"task_id"`
	Manifest string `json:"manifest"`
}

type TaskKubeRunJob struct {
	TaskId   string `json:"task_id"`
	Manifest string `json:"manifest"`
}

type TaskGet struct {
	TaskId string `json:"task_id"`
}

type TaskStreamLogs struct {
	TaskId   string `json:"task_id"`
	Duration uint   `json:"duration"`
	Period   uint   `json:"period"`
}

type Error struct {
	Code    uint   `json:"code"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

type Response struct {
	Succeed bool  `json:"succeed"`
	Error   Error `json:"error"`
}

type Log struct {
	Timestamp time.Time `json:"timestamp"`
	Lines     string    `json:"lines"`
}

type Task struct {
	Id      string `json:"id"`
	Status  string `json:"status"`
	Succeed bool   `json:"succeed"`
	Error   Error  `json:"error"`
	Journal []Log  `json:"journal"`
	Logs    []Log  `json:"logs"`
}

type TaskLogs struct {
	TaskId string `json:"task_id"`
	Logs   []Log  `json:"logs"`
}

type StreamLogs struct {
	StreamId string `json:"stream_id"`
	Logs     []Log  `json:"logs"`
}
