// Copyright 2026 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package sitesetting

import (
	"sync/atomic"

	"github.com/alimy/tryst/cfg"
	"github.com/rocboss/paopao-ce/internal/conf"
)

// RuntimePolicy contains settings that business handlers need to read while
// the service is running. A complete immutable snapshot is published for every
// update so readers never observe a partially updated policy.
type RuntimePolicy struct {
	AllowUserRegister bool
	AllowPhoneBind    bool
}

var currentRuntimePolicy atomic.Pointer[RuntimePolicy]

// CurrentRuntimePolicy returns a copy of the currently effective runtime
// policy. The policy is initialized lazily for package-level consumers used in
// tests, while the normal server startup initializes it through NewService.
func CurrentRuntimePolicy() RuntimePolicy {
	initializeRuntimePolicy()
	return *currentRuntimePolicy.Load()
}

func initializeRuntimePolicy() {
	if currentRuntimePolicy.Load() != nil {
		return
	}
	initial := runtimePolicyFromBootstrap()
	currentRuntimePolicy.CompareAndSwap(nil, &initial)
}

func runtimePolicyFromBootstrap() RuntimePolicy {
	policy := RuntimePolicy{}
	if conf.WebProfileSetting != nil {
		policy.AllowUserRegister = conf.WebProfileSetting.AllowUserRegister
		policy.AllowPhoneBind = conf.WebProfileSetting.AllowPhoneBind && cfg.If("Sms")
	}
	// Preserve the legacy feature as a first-start bootstrap default. Once the
	// policy has been initialized, handlers only consult this atomic snapshot.
	if cfg.If("Web:DisallowUserRegister") {
		policy.AllowUserRegister = false
	}
	return policy
}

func updateRuntimePolicy(update func(*RuntimePolicy)) {
	initializeRuntimePolicy()
	for {
		current := currentRuntimePolicy.Load()
		next := *current
		update(&next)
		if currentRuntimePolicy.CompareAndSwap(current, &next) {
			return
		}
	}
}

func setAllowUserRegister(value bool) {
	updateRuntimePolicy(func(policy *RuntimePolicy) {
		policy.AllowUserRegister = value
	})
}

func setAllowPhoneBind(value bool) {
	updateRuntimePolicy(func(policy *RuntimePolicy) {
		policy.AllowPhoneBind = value && cfg.If("Sms")
	})
}

func resetRuntimePolicyForTest() {
	currentRuntimePolicy.Store(nil)
}
