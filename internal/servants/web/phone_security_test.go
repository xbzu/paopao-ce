package web

import (
	"testing"

	modelweb "github.com/rocboss/paopao-ce/internal/model/web"
)

func TestPhoneOperationsRejectWhenSMSVerificationIsDisabled(t *testing.T) {
	previous := _enablePhoneVerify
	_enablePhoneVerify = false
	t.Cleanup(func() { _enablePhoneVerify = previous })

	if err := (&pubSrv{}).SendCaptcha(&modelweb.SendCaptchaReq{}); err != modelweb.ErrDisallowPhoneBind {
		t.Fatalf("SendCaptcha() error = %v, want ErrDisallowPhoneBind", err)
	}
	if err := (&coreSrv{}).UserPhoneBind(&modelweb.UserPhoneBindReq{}); err != modelweb.ErrDisallowPhoneBind {
		t.Fatalf("UserPhoneBind() error = %v, want ErrDisallowPhoneBind", err)
	}
}
