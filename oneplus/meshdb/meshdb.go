package meshdb

import (
	"fmt"

	"math/big"
	//"time"

	"github.com/pmker/onegw/oneplus/db"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	//log "github.com/sirupsen/logrus"
)

// MiniHeader is the database representation of a succinct Ethereum gateway headers
type MiniHeader struct {
	Hash   common.Hash
	Parent common.Hash
	Number *big.Int
	Logs   []types.Log
}

type ContractWhiteAddress struct {

}

// ID returns the MiniHeader's ID
func (m *MiniHeader) ID() []byte {
	return m.Hash.Bytes()
}
//
//// Order is the database representation a 0x order along with some relevant metadata
//type Order struct {
//	Hash        tool.Hash
//	SignedOrder *zeroex.SignedOrder
//	// When was this order last validated
//	LastUpdated time.Time
//	// How much of this order can still be filled
//	FillableTakerAssetAmount *big.Int
//	// Was this order flagged for removal? Due to the possibility of gateway-reorgs, instead
//	// of immediately removing an order when FillableTakerAssetAmount becomes 0, we instead
//	// flag it for removal. After this order isn't updated for X time and has IsRemoved = true,
//	// the order can be permanently deleted.
//	IsRemoved bool
//}
//
//// ID returns the Order's ID
//func (o Order) ID() []byte {
//	return o.Hash.Bytes()
//}

// MeshDB instantiates the DB connection and creates all the collections used by the application
type MeshDB struct {
	database    *db.DB
	MiniHeaders *MiniHeadersCollection
	//Orders      *OrdersCollection
}

// MiniHeadersCollection represents a DB collection of mini Ethereum gateway headers
type MiniHeadersCollection struct {
	*db.Collection
	numberIndex *db.Index
}
//
//// OrdersCollection represents a DB collection of 0x orders
//type OrdersCollection struct {
//	*db.Collection
//	MakerAddressAndSaltIndex             *db.Index
//	MakerAddressTokenAddressTokenIDIndex *db.Index
//	LastUpdatedIndex                     *db.Index
//	IsRemovedIndex                       *db.Index
//}

// NewMeshDB instantiates a new MeshDB instance
func NewMeshDB(path string) (*MeshDB, error) {
	database, err := db.Open(path)
	if err != nil {
		return nil, err
	}

	miniHeaders, err := setupMiniHeaders(database)
	if err != nil {
		return nil, err
	}

	//orders, err := setupOrders(database)
	//if err != nil {
	//	return nil, err
	//}

	return &MeshDB{
		database:    database,
		MiniHeaders: miniHeaders,
		//Orders:      orders,
	}, nil
}


func setupMiniHeaders(database *db.DB) (*MiniHeadersCollection, error) {
	col, err := database.NewCollection("miniHeader", &MiniHeader{})
	if err != nil {
		return nil, err
	}
	numberIndex := col.AddIndex("number", func(model db.Model) []byte {
		// By default, the index is sorted in byte order. In order to sort by
		// numerical order, we need to pad with zeroes. The maximum length of an
		// unsigned 256 bit integer is 80, so we pad with zeroes such that the
		// length of the number is always 80.
		number := model.(*MiniHeader).Number
		return []byte(fmt.Sprintf("%080s", number.String()))
	})

	return &MiniHeadersCollection{
		Collection:  col,
		numberIndex: numberIndex,
	}, nil
}

// Close closes the database connection
func (m *MeshDB) Close() {
	m.database.Close()
}

// FindAllMiniHeadersSortedByNumber returns all MiniHeaders sorted by gateway number
func (m *MeshDB) FindAllMiniHeadersSortedByNumber() ([]*MiniHeader, error) {
	miniHeaders := []*MiniHeader{}
	query := m.MiniHeaders.NewQuery(m.MiniHeaders.numberIndex.All())
	err := query.Run(&miniHeaders)
	if err != nil {
		return nil, err
	}
	return miniHeaders, nil
}

// FindLatestMiniHeader returns the latest MiniHeader (i.e. the one with the
// largest gateway number), or nil if there are none in the database.
func (m *MeshDB) FindLatestMiniHeader() (*MiniHeader, error) {
	miniHeaders := []*MiniHeader{}
	query := m.MiniHeaders.NewQuery(m.MiniHeaders.numberIndex.All()).Reverse().Max(1)
	err := query.Run(&miniHeaders)
	if err != nil {
		return nil, err
	}
	if len(miniHeaders) == 0 {
		return nil, nil
	}
	return miniHeaders[0], nil
}


type singleAssetData struct {
	Address common.Address
	TokenID *big.Int
}

