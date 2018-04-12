/*
 * Copyright (C) 2017-2018 Alibaba Group Holding Limited
 */
package openapi

import (
	"github.com/aliyun/aliyun-cli/cli"
	"time"
	"fmt"
	"github.com/aliyun/aliyun-cli/i18n"
	"strconv"
)

var WaiterFlag = &cli.Flag {Category:"helper",
	Name: "waiter",
	AssignedMode: cli.AssignedRepeatable,
	Short: i18n.T(
		"use `--waiter expr=<jmesPath> to=<value>` to pull api until result equal to expected value",
		"使用 `--waiter expr=<jmesPath> to=<value>` 来轮询调用OpenAPI，知道返回期望的值"),
	Long: i18n.T(
		"",
		""),
	Fields: []cli.Field {
		{Key:"expr", Required:true, Short: i18n.T("", "")},
		{Key:"to", Required:true, Short: i18n.T("", "")},
		{Key:"timeout",DefaultValue:"180", Short:i18n.T("", "")},
		{Key:"interval",DefaultValue:"5", Short:i18n.T("","")},
	},
	ExcludeWith:[]string{"pager"},
}

type Waiter struct {
	expr string
	to string
//	timeout  time.Duration	TODO use Flag.Field to validate
//	interval time.Duration  TODO use Flag.Field to validate
}

func GetWaiter() *Waiter {
	if !WaiterFlag.IsAssigned() {
		return nil
	}

	waiter := &Waiter {
	}
	waiter.expr, _ = WaiterFlag.GetFieldValue("expr")
	waiter.to, _ = WaiterFlag.GetFieldValue("to")
	//waiter.timeout = time.Duration(time.Second * 180)
	//waiter.interval = time.Duration(time.Second * 5)

	return waiter
}

func (a *Waiter) CallWith(invoker Invoker) (string, error) {
	//
	// timeout is 1-180 seconds
	timeout := time.Duration(time.Second * 180)
	if s, ok := WaiterFlag.GetFieldValue("timeout"); ok {
		if n, err := strconv.Atoi(s); err == nil {
			if n <= 0 && n > 180 {
				return "", fmt.Errorf("--waiter timeout=%s must between 1-180 (seconds)", s)
			}
			timeout = time.Duration(time.Second * time.Duration(n))
		} else {
			return "", fmt.Errorf("--waiter timeout=%s must be integer", s)
		}
	}
	//
	// interval is 2-10 seconds
	interval := time.Duration(time.Second * 5)
	if s, ok := WaiterFlag.GetFieldValue("interval"); ok {
		if n, err := strconv.Atoi(s); err == nil {
			if n <= 1 && n > 10 {
				return "", fmt.Errorf("--waiter interval=%s must between 2-10 (seconds)", s)
			}
			interval = time.Duration(time.Second * time.Duration(n))
		} else {
			return "", fmt.Errorf("--waiter interval=%s must be integer", s)
		}
	}

	begin := time.Now()
	for {
		resp, err := invoker.Call()
		if err != nil {
			return "", err
		}

		v, err := evaluateExpr(resp.GetHttpContentBytes(), a.expr)
		if err != nil {
			return "", err
		}

		if v == a.to {
			return resp.GetHttpContentString(), nil
		}
		duration := time.Now().Sub(begin)
		if duration > timeout {
			return "", fmt.Errorf("wait '%s' to '%s' timeout(%dseconds), last='%s'",
				a.expr, a.to, timeout / time.Second, v)
		}
		time.Sleep(interval)
	}
}