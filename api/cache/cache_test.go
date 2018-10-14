package cache

import (
	"testing"
  "encoding/json"
)

func TestCache(t *testing.T) {
  cache := New("redis","127.0.0.1:6379", "r7lebWzDLzOC1E0Cv8OM")
  //strarr := []string{"aaa","bbb"}
  str, err := cache.Get("test1")
  //var sss []byte
  //str,err := json.Marshal(strarr)
  //err = cache.Set("test", string(str))
  //t.Logf(str)
  //var strarr []string
  if err != nil {
    t.Logf("err: %v", err)
  } else {
    var strarr []string
    json.Unmarshal(str.([]byte),&strarr)
    t.Fatalf("str: %v\n", strarr[0])
  }
}
