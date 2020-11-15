package ab

import (
	"fmt"
	"sort"

	"github.com/spaolacci/murmur3"
)

const (
	CONTROL = "control" // 对照组实验的名字
)

type Domain map[string]*Layer

type Layer struct {
	Observations map[string]*Experiment `json:"observations"` // 观察组
	Control      *Experiment            `json:"control"`      // 对照组
	WhiteList    map[int64]string       `json:"white_list"`   // 白名单，指定用户走指定实验
	exps         []string               // 所有观察组实验名称，用于构建有序map
	bucketNum    uint32                 // 已使用的总桶数
	name         string                 // 实验层名称
	hashFunc     HashFunc               // 哈希函数
}

type Experiment struct {
	*Layer
	Model  string                 `json:"model"`  // 模型名
	Ratio  uint32                 `json:"ratio"`  // 比例
	Params map[string]interface{} `json:"params"` // 模型参数
	bucket *Bucket                // 流量桶
	name   string                 // 实验名称
}

type Bucket struct {
	offset uint32 // 流量桶起始数字
	size   uint32 // 流量桶大小
}

func (b *Bucket) isEmpty() bool {
	return b.size == 0
}

// id ?\in [offset, offset+size-1]
func (b *Bucket) Contain(id uint32) bool {
	if b.isEmpty() {
		return false
	}
	if id >= b.offset && id < b.offset+b.size {
		return true
	}
	return false
}

func (b *Bucket) String() string {
	return fmt.Sprintf("[%d, %d]", b.offset, b.offset+b.size-1)
}

type HashFunc func(uid int64) uint32

func (d Domain) InitDomain(funcs map[string]HashFunc) {
	for name, layer := range d {
		layer.name = name
		layer.initLayer(funcs[name])
	}
}

func NewDomain() Domain { return Domain{} }

func (d Domain) String() string {
	ret := ""
	for _, layer := range d {
		ret += layer.String()
		ret += "\n"
		for _, exp := range layer.exps {
			ret += "  " + layer.Observations[exp].String() + "\n"
		}
		ret += "  " + layer.Control.String() + "\n"
		ret += "\n"
	}
	return ret
}

func (d Domain) HasLayer(layer string) bool {
	if _, ok := d[layer]; !ok {
		return false
	}
	return true
}

func (d Domain) GetLayer(layer string) *Layer {
	if d.HasLayer(layer) {
		return d[layer]
	}
	return nil
}

// 根据uid和layer，拿到对应layer的实验
func (d Domain) GetExperiment(uid int64, layer string) (*Experiment, error) {
	l := d.GetLayer(layer)
	if l == nil {
		return nil, fmt.Errorf("GetExperiment failed: layer %s not exists", layer)
	}
	// 优先根据白名单分配实验
	if exp := l.getWhiteListExperiment(uid); exp != nil {
		return exp, nil
	}
	if l.GetBucketNum() == 0 { // 未给任何实验分配流量桶，默认走control
		return l.Control, nil
	}
	id := l.GetBucketID(uid)
	for _, exp := range l.Observations {
		if exp.bucket.Contain(id) {
			return exp, nil
		}
	}
	return l.Control, nil
}

// 检查domain中是否存在指定的layer
func (d Domain) CheckLayers(layers []string) error {
	for _, layer := range layers {
		if _, ok := d[layer]; !ok {
			return fmt.Errorf("CheckLayer faild: layer %s not exist", layer)
		}
	}
	return nil
}

func (l *Layer) initLayer(f HashFunc) {
	// 初始化指定hash函数
	l.hashFunc = f
	if l.hashFunc == nil {
		l.hashFunc = l.DefaultHashFunc
	}
	for name := range l.Observations {
		l.exps = append(l.exps, name)
	}
	// 按实验名称排序，按顺序遍历map（有序字典）
	sort.Slice(l.exps, func(i, j int) bool {
		return l.exps[i] < l.exps[j]
	})
	// 初始化观察组实验
	for _, name := range l.exps {
		l.Observations[name].name = name
		l.Observations[name].initExperiment(l)
	}
	// 初始化对照组实验
	l.Control.name = CONTROL
	l.Control.initExperiment(l)
}

func (l *Layer) GetName() string { return l.name }

func (l *Layer) DefaultHashFunc(uid int64) uint32 {
	// 联合uid和layer作为code
	code := fmt.Sprintf("%d_%s", uid, l.GetName()) // uid_layer
	// 使用murmur3对code做hash，hash值对应的桶，对应相应的实验
	id := murmur3.Sum32(([]byte)(code))
	return id
}

func (l *Layer) GetHash(uid int64) uint32 { return l.hashFunc(uid) }

// 需要先initBucketNum
func (l *Layer) GetBucketNum() uint32 { return l.bucketNum }

func (l *Layer) GetBucketID(uid int64) uint32 { return l.GetHash(uid) % l.GetBucketNum() }

func (l *Layer) getWhiteListExperiment(uid int64) *Experiment {
	name, ok := l.WhiteList[uid]
	if !ok {
		return nil
	}
	if name == CONTROL {
		return l.Control
	}
	if exp, ok := l.Observations[name]; ok {
		return exp
	}
	return nil
}

func (l *Layer) String() string {
	ret := fmt.Sprintf("[Layer]:%-8s whiteList: %v", l.GetName(), l.WhiteList)
	return ret
}

func (e *Experiment) GetName() string { return e.name }

func (e *Experiment) GetBucket() *Bucket { return e.bucket }

func (e *Experiment) initExperiment(layer *Layer) {
	if e.Layer == nil {
		e.Layer = layer
	}
	e.initBucket(e.bucketNum)
}

func (e *Experiment) initBucket(offset uint32) {
	e.bucket = &Bucket{offset, e.Ratio}
	e.bucketNum += e.Ratio
}

func (e *Experiment) String() string {
	ret := fmt.Sprintf("[exp]: %-8s  layer: %-8s  model: %-12s  bucket: %-12v  params: %s", e.GetName(), e.Layer.GetName(),
		e.Model, e.GetBucket(), e.Params)
	return ret
}
