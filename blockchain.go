package blockchain

import (
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const blocksBucket = "blocksBucket"
const dbFile = "dbFile"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

//Blockchain type.  Consists of a pointer to a db and a tip.  The tip is the top of the chain
type Blockchain struct {
	tip []byte
	Db  *bolt.DB
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (blockchain *Blockchain) Iterator() *BlockchainIterator {
	blockchainIterator := &BlockchainIterator{blockchain.tip, blockchain.Db}
	return blockchainIterator
}

func (blockchainIterator *BlockchainIterator) Next() *Block {
	var block *Block
	err := blockchainIterator.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		encodedBlock := bucket.Get(blockchainIterator.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	blockchainIterator.currentHash = block.PrevBlockHash

	return block

}

func (blockchain *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	err := blockchain.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		lastHash = bucket.Get([]byte("1"))

		return nil
	})

	newBlock := NewBlock(transactions, lastHash)

	err = blockchain.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		err := bucket.Put(newBlock.Hash, newBlock.Serialize())
		err = bucket.Put([]byte("1"), newBlock.Hash)
		blockchain.tip = newBlock.Hash

		if err != nil {
			log.Fatal(err)
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

//Creates a new blockchain and adds the genesis block
//probably deprecated... passing in address makes no sense...
func NewBlockChain(address string) *Blockchain {
	var tip []byte

	if !dbExists() {
		fmt.Println("No existing blockchain found.  Create one first.")
		os.Exit(1)
	}

	db, err := bolt.Open(dbFile, 0600, nil)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("1"))
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	blockchain := Blockchain{tip, db}

	return &blockchain

}

func CreateBlockChain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(transaction *bolt.Tx) error {
		coinbaseTransaction := NewCoinbaseTransaction(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(coinbaseTransaction)

		b, err := transaction.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("1"), genesis.Hash)

		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash
		return nil

	})

	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{tip, db}
}
