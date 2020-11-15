# ABTest
This is a simple abTest Golang module designed for testing recommendation algorithms.

It can be used as an interceptor in the web server. 

ABTest can be configured dynamically by watching key from *etcd*.


## Three Entities in ABTest

* Domain: including layers, e.g., [recall layer, rank layer].

* Layer: including experiments. Experiments are mutual inside a layer, they are orthogonal between layers.

* Experiment: experiment entity.



## Usage
In `example/rec_domain.go`, we will new a domain for our application. Hash function for a certain layer is custom.
Default hash func is murmur3 hash. In addition, it supports whiteList in a layer.
```go
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
```
Generate a new domain, and read domain config from etcd:
```go
func main() {
    rawJson := "......" // reads from etcd
    err := UpdateDomain(([]byte)(rawJson))
        if err != nil {
            fmt.Println(err)
        }
        domain := CurrentDomain()
        fmt.Println(domain)
}
```
It will print as:
```shell
[Layer]:recall   whiteList: map[243542355:exp1 12343344:control]
  [exp]: exp1      layer: recall    model: als           bucket: [0, 14]
  [exp]: exp2      layer: recall    model: i2i           bucket: [14, 45]
  [exp]: exp3      layer: recall    model: embedding     bucket: [45, 66]
  [exp]: control   layer: recall    model: als           bucket: [66, 83]

[Layer]:rank     whiteList: map[]
  [exp]: exp1      layer: rank      model: deepfm        bucket: [0, 24]
  [exp]: control   layer: rank      model: wd_v1         bucket: [24, 58]

```
For different users
```go
for _, uid := range uids {
    for _, layer := range layers {
        e, err := domain.GetExperiment(uid, layer)
        if err == nil {
            fmt.Println(uid, domain.GetLayer(layer).GetBucketID(uid), e)
        }
    }
    fmt.Println()
}
```
It will get
```shell script
243542355 5 [exp]: exp1      layer: recall    model: als           ratio:1         bucket: [0, 0]
243542355 52 [exp]: control   layer: rank      model: wd_v1         ratio:34        bucket: [24, 57]

243542351 1 [exp]: exp2      layer: recall    model: i2i           ratio:2         bucket: [1, 2]
243542351 20 [exp]: exp1      layer: rank      model: deepfm        ratio:24        bucket: [0, 23]

243542353 3 [exp]: exp3      layer: recall    model: embedding     ratio:2         bucket: [3, 4]
243542353 38 [exp]: control   layer: rank      model: wd_v1         ratio:34        bucket: [24, 57]
```