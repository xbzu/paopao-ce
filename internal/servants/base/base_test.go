package base

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rocboss/paopao-ce/internal/model/joint"
	"github.com/rocboss/paopao-ce/pkg/xerror"
)

func TestRenderersPreserveValidationDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		render func(*gin.Context, error)
	}{
		{name: "RenderAny", render: func(ctx *gin.Context, err error) { RenderAny(ctx, nil, err) }},
		{name: "BaseServant.Render", render: func(ctx *gin.Context, err error) { (&BaseServant{}).Render(ctx, nil, err) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			test.render(ctx, xerror.InvalidParams.WithDetails("setting feature is not active"))

			var response joint.JsonResp
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response.Code != xerror.InvalidParams.StatusCode() {
				t.Fatalf("code = %d, want %d", response.Code, xerror.InvalidParams.StatusCode())
			}
			if len(response.Details) != 1 || response.Details[0] != "setting feature is not active" {
				t.Fatalf("details = %v", response.Details)
			}
		})
	}
}

func TestBindValidationDetailsSurviveRendering(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("POST", "/", nil)
	request := struct {
		Name string `form:"name" binding:"required"`
	}{}
	err := bindAny(ctx, &request)
	if err == nil {
		t.Fatal("bindAny() error = nil, want validation error")
	}
	RenderAny(ctx, nil, err)

	var response joint.JsonResp
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Code != xerror.InvalidParams.StatusCode() || len(response.Details) != 1 || response.Details[0] == "" {
		t.Fatalf("response = code:%d details:%v", response.Code, response.Details)
	}
}
