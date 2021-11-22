package main

import (
	"fmt"
	"github.com/goodsign/monday"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestToString(t *testing.T) {
	pfd := ParamFillDate{
		Attribute: "today",
		Format:    "2006-02-01",
	}
	require.Equal(t, "тест: "+monday.Format(time.Now(), pfd.Format, monday.LocaleRuRU), fmt.Sprintf("тест: %s", &pfd))
}
