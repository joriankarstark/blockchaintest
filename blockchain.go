package blockchain

import (
	"log"

	"github.com/boltdb/bolt"
)

const blocksBucket = "blocksBucket"

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (blockchain *Blockchain) Iterator() *BlockchainIterator {
	blockchainIterator := &BlockchainIterator{blockchain.tip, blockchain.db}
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

func (blockchain *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		lastHash = bucket.Get([]byte("1"))

		return nil
	})

	newBlock := NewBlock(data, lastHash)

	err = blockchain.db.Update(func(tx *bolt.Tx) error {
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

func NewBlockChain() *Blockchain {
	var tip []byte
	db, err := bolt.Open("dbFile", 0600, nil)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			b.Put(genesis.Hash, genesis.Serialize())
			err = b.Put([]byte("1"), genesis.Hash)
			tip = genesis.Hash

			if err != nil {
				log.Fatal(err)
			}

		} else {
			tip = b.Get([]byte("1"))
		}
		return nil
	})

	blockchain := Blockchain{tip, db}

	return &blockchain

}