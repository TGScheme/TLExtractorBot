package ast_parser

import (
	"fmt"
	"math"
	"strconv"
)

type flagContext struct {
	flagFields map[string]bool
	flagOrder  []string
	boolToBit  map[string][2]string
	boolGuards map[string]bool
}

func newFlagContext() *flagContext {
	return &flagContext{
		flagFields: make(map[string]bool),
		boolToBit:  make(map[string][2]string),
		boolGuards: make(map[string]bool),
	}
}

func (f *flagContext) declareFlagField(name string) {
	if f.flagFields[name] {
		return
	}
	f.flagFields[name] = true
	f.flagOrder = append(f.flagOrder, name)
}
func (f *flagContext) isFlagField(name string) bool { return f.flagFields[name] }

func (f *flagContext) firstFlagField() string {
	if len(f.flagOrder) == 0 {
		return ""
	}
	return f.flagOrder[0]
}

func (f *flagContext) mapBool(boolVar, flagField string, bit int) {
	f.boolToBit[boolVar] = [2]string{flagField, strconv.Itoa(bit)}
}

func (f *flagContext) lookupBool(boolVar string) (string, int, bool) {
	v, ok := f.boolToBit[boolVar]
	if !ok {
		return "", 0, false
	}
	bit, _ := strconv.Atoi(v[1])
	return v[0], bit, true
}

func payloadKey(flagField string, bit int) string { return fmt.Sprintf("%s#%d", flagField, bit) }

func (f *flagContext) markBoolGuard(flagField string, bit int) {
	f.boolGuards[payloadKey(flagField, bit)] = true
}
func (f *flagContext) isBoolGuarded(flagField string, bit int) bool {
	return f.boolGuards[payloadKey(flagField, bit)]
}

func flagPrefix(flagField string, bit int, tlType string) string {
	return fmt.Sprintf("%s.%d?%s", flagField, bit, tlType)
}

func bitOfMask(mask string) (int, bool) {
	n, err := strconv.Atoi(mask)
	if err != nil || n <= 0 {
		return 0, false
	}
	if n&(n-1) != 0 {
		return 0, false
	}
	return int(math.Log2(float64(n))), true
}
