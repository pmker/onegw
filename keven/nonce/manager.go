package nonce

import (
	"github.com/pmker/onegw/oneplus/backend/sdk/ethereum"
	"math/big"
	"sync"
)

// 管理器结构体
type NonceManager struct {
	// lock 是互斥锁，go 的 map 类型不是线程安全的，
	// 在读写 map 的时候，我们要考虑上多协程并发的情况
	lock sync.Mutex
	// 采用整形大数来存储 nonce
	nonceMemCache map[string]*big.Int

	Hydro *ethereum.EthereumHydro
}

func NewNonceManager(hydro *ethereum.EthereumHydro) *NonceManager {
	return &NonceManager{
		lock:  sync.Mutex{}, // 实例化互斥锁
		Hydro: hydro,
	}
}

//  设置 nonce
func (n *NonceManager) SetNonce(address string, nonce *big.Int) {
	if n.nonceMemCache == nil {
		n.nonceMemCache = map[string]*big.Int{}
	}
	n.lock.Lock()         // 加锁
	defer n.lock.Unlock() // 当该函数执行完毕，进行解锁
	n.nonceMemCache[address] = nonce
}

//  同步 nonce
func (n *NonceManager) SyncNonce(address string) (*big.Int, error) {
	if n.nonceMemCache == nil {
		n.nonceMemCache = map[string]*big.Int{}
	}
	n.lock.Lock()         // 加锁
	defer n.lock.Unlock() // 当该函数执行完毕，进行解锁

	// nonce 不存在，开始访问节点获取
	n1, err := n.Hydro.GetTransactionCount(address)
	nonce := new(big.Int).SetUint64(uint64(n1))

	n.nonceMemCache[address] = nonce

	///go n.SetNonce(address,nonce)
	////n2:=n.GetNonce(address)
	//
	//fmt.Printf("Address:%s,nonce11,%d,%s,%d", address, n1, nonce)
	//
	////n2:=n.GetNonce(address)
	////fmt.Printf("Address:%s,nonce,%d,%s", address, n2, nonce)
	//
	//if err != nil {
	//	fmt.Errorf("获取 nonce 失败 %s", err.Error())
	//}
	return nonce, err

}

// 根据以太坊地址获取 nonce
func (n *NonceManager) GetNonceOk(address string) (*big.Int, error) {
	if n.nonceMemCache == nil {
		n.nonceMemCache = map[string]*big.Int{}
	}
	n.lock.Lock()         // 加锁
	defer n.lock.Unlock() // 当该函数执行完毕，进行解锁
	nonce := n.nonceMemCache[address]
	if nonce == nil {
		n1, err := n.Hydro.GetTransactionCount(address)
		if err != nil {
			return nil, err
		}

		//fmt.Printf("\n get nonce ok chian %d ", n1)

		nonce := new(big.Int).SetUint64(uint64(n1))
		//newNonce:=nonce.Add(nonce, big.NewInt(int64(1)))
		//fmt.Printf("\n get nonce ok chian %d,%d  \n", n1, nonce)
		 n.nonceMemCache[address] = nonce
		return nonce, nil
	}
	return nonce, nil
}

// 根据以太坊地址获取 nonce
func (n *NonceManager) GetNonce(address string) *big.Int {
	if n.nonceMemCache == nil {
		n.nonceMemCache = map[string]*big.Int{}
	}
	n.lock.Lock()         // 加锁
	defer n.lock.Unlock() // 当该函数执行完毕，进行解锁
	nonce := n.nonceMemCache[address]

	return nonce
}

// nonce 进行自增 1
func (n *NonceManager) PlusNonce(address string) {



	if n.nonceMemCache == nil {
		n.nonceMemCache = map[string]*big.Int{}
	}
	n.lock.Lock()         // 加锁
	defer n.lock.Unlock() // 当该函数执行完毕，进行解锁
	oldNonce := n.nonceMemCache[address]
	//fmt.Printf("\n fiuckafadf address:%s,%d \n",address,oldNonce )
	//for k,v := range n.nonceMemCache {
	//	fmt.Printf("\n adfasd %s,%v \n",k,v)
	//}

	if oldNonce == nil {
		return
	}
	newNonce := oldNonce.Add(oldNonce, big.NewInt(int64(1)))

	//fmt.Printf("\n add address:%s,%d \n",address,newNonce)
	n.nonceMemCache[address] = newNonce
}
