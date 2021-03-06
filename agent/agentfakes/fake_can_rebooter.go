// Code generated by counterfeiter. DO NOT EDIT.
package agentfakes

import (
	sync "sync"

	agent "github.com/cloudfoundry/bosh-agent/agent"
)

type FakeCanRebooter struct {
	CanRebootStub        func() (bool, error)
	canRebootMutex       sync.RWMutex
	canRebootArgsForCall []struct {
	}
	canRebootReturns struct {
		result1 bool
		result2 error
	}
	canRebootReturnsOnCall map[int]struct {
		result1 bool
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCanRebooter) CanReboot() (bool, error) {
	fake.canRebootMutex.Lock()
	ret, specificReturn := fake.canRebootReturnsOnCall[len(fake.canRebootArgsForCall)]
	fake.canRebootArgsForCall = append(fake.canRebootArgsForCall, struct {
	}{})
	fake.recordInvocation("CanReboot", []interface{}{})
	fake.canRebootMutex.Unlock()
	if fake.CanRebootStub != nil {
		return fake.CanRebootStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.canRebootReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeCanRebooter) CanRebootCallCount() int {
	fake.canRebootMutex.RLock()
	defer fake.canRebootMutex.RUnlock()
	return len(fake.canRebootArgsForCall)
}

func (fake *FakeCanRebooter) CanRebootCalls(stub func() (bool, error)) {
	fake.canRebootMutex.Lock()
	defer fake.canRebootMutex.Unlock()
	fake.CanRebootStub = stub
}

func (fake *FakeCanRebooter) CanRebootReturns(result1 bool, result2 error) {
	fake.canRebootMutex.Lock()
	defer fake.canRebootMutex.Unlock()
	fake.CanRebootStub = nil
	fake.canRebootReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeCanRebooter) CanRebootReturnsOnCall(i int, result1 bool, result2 error) {
	fake.canRebootMutex.Lock()
	defer fake.canRebootMutex.Unlock()
	fake.CanRebootStub = nil
	if fake.canRebootReturnsOnCall == nil {
		fake.canRebootReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 error
		})
	}
	fake.canRebootReturnsOnCall[i] = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeCanRebooter) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.canRebootMutex.RLock()
	defer fake.canRebootMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeCanRebooter) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ agent.CanRebooter = new(FakeCanRebooter)
