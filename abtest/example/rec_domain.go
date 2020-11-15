package main

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"unsafe"

	"github.com/yinyajun/abTest"
)

const (
	RECALL = "recall"
	ROUGH  = "rough"
	RANK   = "rank"
	RERANK = "rerank"
)

var DomainPtr unsafe.Pointer = unsafe.Pointer(&ab.Domain{})

func CurrentDomain() ab.Domain {
	return *(*ab.Domain)(atomic.LoadPointer(&DomainPtr))
}

func UpdateDomain(v []byte) error {
	domain := ab.NewDomain()
	if err := json.Unmarshal(v, &domain); err != nil {
		return err
	}
	layers := []string{RECALL, RANK}
	funcs := map[string]ab.HashFunc{RECALL: hashByTail} // custom hash func for certain layer
	if err := domain.CheckLayers(layers); err != nil {
		return err
	}
	domain.InitDomain(funcs)
	atomic.StorePointer(&DomainPtr, unsafe.Pointer(&domain))
	return nil
}

func hashByTail(uid int64) uint32 {
	return uint32(uid % 10)
}

func main() {
	j := `{
		  "recall": {
			"observations": {
			  "exp1": {
				"model": "als",
				"ratio": 1,
				"params": {
				  "aa": "ds"
				}
			  },
			  "exp2": {
				"model": "i2i",
				"params": {},
				"ratio": 2
			  },
			  "exp3": {
				"model": "embedding",
				"ratio": 2,
				"params": {
				  "fea": "ge"
				}
			  }
			},
			"control": {
			  "model": "als",
			  "ratio": 5,
			  "params": {
				"cas": "efa"
			  }
			},
			"white_list": {
			  "12343344": "control",
			  "243542355": "exp1"
			}
		  },
		  "rank": {
			"observations": {
			  "exp1": {
				"model": "deepfm",
				"params": {
				  "fea": "523"
				},
				"ratio": 24
			  }
			},
			"control": {
			  "model": "wd_v1",
			  "params": {
				"fea": "432"
			  },
			  "ratio": 34
			},
			"white_list": {}
		  }
		}`

	err := UpdateDomain(([]byte)(j))
	if err != nil {
		fmt.Println(err)
	}
	domain := CurrentDomain()
	fmt.Println(domain)

	uids := []int64{243542355, 243542351, 243542353}
	layers := []string{RECALL, RANK}
	for _, uid := range uids {
		for _, layer := range layers {
			e, err := domain.GetExperiment(uid, layer)
			if err == nil {
				fmt.Println(uid, domain.GetLayer(layer).GetBucketID(uid), e)
			}
		}
		fmt.Println()
	}
}
