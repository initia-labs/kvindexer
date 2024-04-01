package dashboard

import (
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/sha3"
	"time"
)

const dateFmt = "2006-01-02"

func timeToDateString(t time.Time) string {
	return t.Format(dateFmt)
}

func getOpDenom(bridgeID uint64, l1Denom string) string {
	bridgeIDBuf := make([]byte, 8)
	for i := 0; i < 8; i++ {
		bridgeIDBuf[7-i] = byte(bridgeID >> (i * 8))
	}
	hash := sha3.Sum256(append(bridgeIDBuf, []byte(l1Denom)...))

	return fmt.Sprintf("l2/%s", hex.EncodeToString(hash[:]))
}
