package eth_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/go-ethlibs/eth"
)

func TestLogFilter_Matches(t *testing.T) {
	makeLog := func(address eth.Address, topics []eth.Topic, number uint64, hashes ...*eth.Hash) eth.Log {
		bn := eth.QuantityFromUInt64(number)
		var hash, txHash *eth.Hash
		if len(hashes) > 0 {
			hash = hashes[0]
		} else {
			// Compute a hash from the block number
			h, _ := bn.RLP().Hash()
			hash = eth.MustHash(h)
		}
		if len(hashes) > 1 {
			txHash = hashes[1]
		} else {
			// Compute a hash from the block hash
			t, _ := address.RLP().Hash()
			txHash = eth.MustHash(t)
		}

		return eth.Log{
			Removed:     false,
			LogIndex:    eth.MustQuantity("0x1"),
			TxIndex:     eth.MustQuantity("0x2"),
			TxHash:      txHash,
			BlockHash:   hash,
			BlockNumber: &bn,
			Address:     address,
			Data:        *eth.MustData("0x0102"),
			Topics:      topics,
		}
	}

	t.Run("Addresses", func(t *testing.T) {
		t.Run("EmptyFilterMatchesAddress", func(t *testing.T) {
			f := eth.LogFilter{}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.True(t, f.Matches(makeLog(addr, []eth.Topic{}, 1234)))
		})

		t.Run("SingleMatch", func(t *testing.T) {
			addr1 := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			addr2 := *eth.MustAddress("0x2222222222222222222222222222222222222222")
			f := eth.LogFilter{
				Address: []eth.Address{
					addr1,
				},
			}
			require.True(t, f.Matches(makeLog(addr1, []eth.Topic{}, 1234)))
			require.False(t, f.Matches(makeLog(addr2, []eth.Topic{}, 1234)))
		})

		t.Run("MultiMatch", func(t *testing.T) {
			addr1 := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			addr2 := *eth.MustAddress("0x2222222222222222222222222222222222222222")
			addr3 := *eth.MustAddress("0x3333333333333333333333333333333333333333")
			f := eth.LogFilter{
				Address: []eth.Address{
					addr1,
					addr2,
				},
			}
			require.True(t, f.Matches(makeLog(addr1, []eth.Topic{}, 1234)))
			require.True(t, f.Matches(makeLog(addr2, []eth.Topic{}, 1234)))
			require.False(t, f.Matches(makeLog(addr3, []eth.Topic{}, 1234)))
		})
	})

	t.Run("Topics", func(t *testing.T) {
		t.Run("MissingTopic", func(t *testing.T) {
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			f := eth.LogFilter{
				Topics: [][]eth.Topic{
					{
						topic1,
					},
				},
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.False(t, f.Matches(makeLog(addr, []eth.Topic{}, 1234)))
		})

		t.Run("MissingSecondTopic", func(t *testing.T) {
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			f := eth.LogFilter{
				Topics: [][]eth.Topic{
					{},
					{
						topic1,
					},
				},
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.False(t, f.Matches(makeLog(addr, []eth.Topic{topic1}, 1234)))
		})

		t.Run("NoTopics", func(t *testing.T) {
			f := eth.LogFilter{}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.True(t, f.Matches(makeLog(addr, []eth.Topic{}, 1234)))
		})

		t.Run("FilterWithoutTopicMatchesLogWithTopic", func(t *testing.T) {
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			f := eth.LogFilter{}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.True(t, f.Matches(makeLog(addr, []eth.Topic{topic1}, 1234)))
		})

		t.Run("Match", func(t *testing.T) {
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			f := eth.LogFilter{
				Topics: [][]eth.Topic{
					{
						topic1,
					},
				},
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.True(t, f.Matches(makeLog(addr, []eth.Topic{topic1}, 1234)))
		})

		t.Run("MatchOne", func(t *testing.T) {
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			topic2 := *eth.MustTopic("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
			f := eth.LogFilter{
				Topics: [][]eth.Topic{
					{
						topic1,
						topic2,
					},
				},
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.True(t, f.Matches(makeLog(addr, []eth.Topic{topic2}, 1234)))
		})

		t.Run("MismatchOne", func(t *testing.T) {
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			topic2 := *eth.MustTopic("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
			f := eth.LogFilter{
				Topics: [][]eth.Topic{
					{
						topic1,
					},
				},
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.False(t, f.Matches(makeLog(addr, []eth.Topic{topic2}, 1234)))
		})

		t.Run("MismatchAll", func(t *testing.T) {
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			topic2 := *eth.MustTopic("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
			topic3 := *eth.MustTopic("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
			f := eth.LogFilter{
				Topics: [][]eth.Topic{
					{
						topic1,
					},
				},
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.False(t, f.Matches(makeLog(addr, []eth.Topic{topic2, topic3}, 1234)))
		})

		t.Run("AnyFirstAndMatchSecond", func(t *testing.T) {
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			topic2 := *eth.MustTopic("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
			topic3 := *eth.MustTopic("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
			f := eth.LogFilter{
				Topics: [][]eth.Topic{
					nil,
					{
						topic2,
					},
				},
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.True(t, f.Matches(makeLog(addr, []eth.Topic{topic1, topic2, topic3}, 1234)))
		})

		t.Run("AnyFirstAndMismatchSecond", func(t *testing.T) {
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			topic2 := *eth.MustTopic("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
			topic3 := *eth.MustTopic("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
			f := eth.LogFilter{
				Topics: [][]eth.Topic{
					nil,
					{
						topic2,
					},
				},
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.False(t, f.Matches(makeLog(addr, []eth.Topic{topic1, topic3, topic3}, 1234)))
		})

		t.Run("OneOfFirstMatches", func(t *testing.T) {
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			topic2 := *eth.MustTopic("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
			topic3 := *eth.MustTopic("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
			f := eth.LogFilter{
				Topics: [][]eth.Topic{
					{
						topic1, topic3,
					},
				},
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.True(t, f.Matches(makeLog(addr, []eth.Topic{topic1, topic2}, 1234)))
		})

		t.Run("OneOfMatchesBoth", func(t *testing.T) {
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			topic2 := *eth.MustTopic("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
			f := eth.LogFilter{
				Topics: [][]eth.Topic{
					{
						topic1, topic2,
					},
					{
						topic1, topic2,
					},
				},
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.True(t, f.Matches(makeLog(addr, []eth.Topic{topic1, topic2}, 1234)))
		})

		t.Run("FirstMatchesSecondDoesNot", func(t *testing.T) {
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			topic2 := *eth.MustTopic("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
			topic3 := *eth.MustTopic("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
			f := eth.LogFilter{
				Topics: [][]eth.Topic{
					{
						topic1, topic2,
					},
					{
						topic1, topic2,
					},
				},
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.False(t, f.Matches(makeLog(addr, []eth.Topic{topic1, topic3}, 1234)))
		})
	})

	t.Run("Blocks", func(t *testing.T) {
		t.Run("BlockHash", func(t *testing.T) {
			t.Run("Match", func(t *testing.T) {
				bh := eth.MustHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
				f := eth.LogFilter{
					BlockHash: bh,
				}
				addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
				require.True(t, f.Matches(makeLog(addr, []eth.Topic{}, 1234, bh)))
			})

			t.Run("Mismatch", func(t *testing.T) {
				hash1 := eth.MustHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
				hash2 := eth.MustHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")

				f := eth.LogFilter{
					BlockHash: hash1,
				}
				addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
				require.False(t, f.Matches(makeLog(addr, []eth.Topic{}, 1234, hash2)))
			})
		})
	})

	t.Run("FromBlock", func(t *testing.T) {
		t.Run("Equal", func(t *testing.T) {
			bn := eth.MustQuantity("0x1234")
			f := eth.LogFilter{
				FromBlock: eth.MustBlockNumberOrTag(bn.String()),
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.True(t, f.Matches(makeLog(addr, []eth.Topic{}, bn.UInt64())))
		})

		t.Run("Greater", func(t *testing.T) {
			bn := eth.MustQuantity("0x1234")
			f := eth.LogFilter{
				FromBlock: eth.MustBlockNumberOrTag(bn.String()),
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.True(t, f.Matches(makeLog(addr, []eth.Topic{}, bn.UInt64()+1)))
		})

		t.Run("Lesser", func(t *testing.T) {
			bn := eth.MustQuantity("0x1234")
			f := eth.LogFilter{
				FromBlock: eth.MustBlockNumberOrTag(bn.String()),
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.False(t, f.Matches(makeLog(addr, []eth.Topic{}, bn.UInt64()-1)))
		})
	})

	t.Run("ToBlock", func(t *testing.T) {
		t.Run("Equal", func(t *testing.T) {
			bn := eth.MustQuantity("0x1234")
			f := eth.LogFilter{
				ToBlock: eth.MustBlockNumberOrTag(bn.String()),
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.True(t, f.Matches(makeLog(addr, []eth.Topic{}, bn.UInt64())))
		})

		t.Run("Greater", func(t *testing.T) {
			bn := eth.MustQuantity("0x1234")
			f := eth.LogFilter{
				ToBlock: eth.MustBlockNumberOrTag(bn.String()),
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.False(t, f.Matches(makeLog(addr, []eth.Topic{}, bn.UInt64()+1)))
		})

		t.Run("Lesser", func(t *testing.T) {
			bn := eth.MustQuantity("0x1234")
			f := eth.LogFilter{
				ToBlock: eth.MustBlockNumberOrTag(bn.String()),
			}
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			require.True(t, f.Matches(makeLog(addr, []eth.Topic{}, bn.UInt64()-1)))
		})
	})

	t.Run("Composite", func(t *testing.T) {
		t.Run("AllMatch", func(t *testing.T) {
			bn := eth.MustQuantity("0x1234")
			from := eth.MustQuantity("0x1000")
			to := eth.MustQuantity("0x1fff")
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			topic2 := *eth.MustTopic("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
			topic3 := *eth.MustTopic("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
			f := eth.LogFilter{
				FromBlock: eth.MustBlockNumberOrTag(from.String()),
				ToBlock:   eth.MustBlockNumberOrTag(to.String()),
				Address:   []eth.Address{addr},
				Topics: [][]eth.Topic{
					{topic1},
					{topic2},
					{topic3},
				},
			}
			require.True(t, f.Matches(makeLog(addr, []eth.Topic{topic1, topic2, topic3}, bn.UInt64())))
		})
		t.Run("WrongBlock", func(t *testing.T) {
			bn := eth.MustQuantity("0x9999")
			from := eth.MustQuantity("0x1000")
			to := eth.MustQuantity("0x1fff")
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			topic2 := *eth.MustTopic("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
			topic3 := *eth.MustTopic("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
			f := eth.LogFilter{
				FromBlock: eth.MustBlockNumberOrTag(from.String()),
				ToBlock:   eth.MustBlockNumberOrTag(to.String()),
				Address:   []eth.Address{addr},
				Topics: [][]eth.Topic{
					{topic1},
					{topic2},
					{topic3},
				},
			}
			require.False(t, f.Matches(makeLog(addr, []eth.Topic{topic1, topic2, topic3}, bn.UInt64())))
		})
		t.Run("WrongAddresss", func(t *testing.T) {
			bn := eth.MustQuantity("0x1234")
			from := eth.MustQuantity("0x1000")
			to := eth.MustQuantity("0x1fff")
			addr1 := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			addr2 := *eth.MustAddress("0x2222222222222222222222222222222222222222")
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			topic2 := *eth.MustTopic("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
			topic3 := *eth.MustTopic("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
			f := eth.LogFilter{
				FromBlock: eth.MustBlockNumberOrTag(from.String()),
				ToBlock:   eth.MustBlockNumberOrTag(to.String()),
				Address:   []eth.Address{addr1},
				Topics: [][]eth.Topic{
					{topic1},
					{topic2},
					{topic3},
				},
			}
			require.False(t, f.Matches(makeLog(addr2, []eth.Topic{topic1, topic2, topic3}, bn.UInt64())))
		})
		t.Run("WrongTopics", func(t *testing.T) {
			bn := eth.MustQuantity("0x1234")
			from := eth.MustQuantity("0x1000")
			to := eth.MustQuantity("0x1fff")
			addr := *eth.MustAddress("0x1111111111111111111111111111111111111111")
			topic1 := *eth.MustTopic("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			topic2 := *eth.MustTopic("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
			topic3 := *eth.MustTopic("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
			f := eth.LogFilter{
				FromBlock: eth.MustBlockNumberOrTag(from.String()),
				ToBlock:   eth.MustBlockNumberOrTag(to.String()),
				Address:   []eth.Address{addr},
				Topics: [][]eth.Topic{
					{topic3},
					{topic3},
				},
			}
			require.False(t, f.Matches(makeLog(addr, []eth.Topic{topic1, topic2}, bn.UInt64())))
		})
	})
}

func TestLogFilterParsing(t *testing.T) {
	type TestCase struct {
		Message  string
		Payload  string
		Expected eth.LogFilter
		Error    bool
	}

	tests := []TestCase{
		{
			Message: "empty params should parse",
			Payload: `{}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "null params should parse",
			Payload: `{"address":null, "topics":null, "blockHash": null, "fromBlock": null, "toBlock": null}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "single address as a string should be supported",
			Payload: `{"address": "0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c"}`,
			Expected: eth.LogFilter{
				Address:   []eth.Address{*eth.MustAddress("0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c")},
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "single address as an array should be supported",
			Payload: `{"address": ["0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c"]}`,
			Expected: eth.LogFilter{
				Address:   []eth.Address{*eth.MustAddress("0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c")},
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "multiple addresses should be supported",
			Payload: `{"address": ["0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c", "0x5717adf502fd8830456bd5dc26801a4db394e6b2"]}`,
			Expected: eth.LogFilter{
				Address: []eth.Address{
					*eth.MustAddress("0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c"),
					*eth.MustAddress("0x5717adf502fd8830456bd5dc26801a4db394e6b2"),
				},
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "single topic should be supported",
			Payload: `{"topics": ["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"]}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics: [][]eth.Topic{
					{eth.Topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
				},
			},
		},
		{
			Message: "multiple topics should be supported",
			Payload: `{"topics": ["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", "0x000000000000000000000000896dd350806eba53dfa9778c4698224e8ede2c41"]}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics: [][]eth.Topic{
					{eth.Topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
					{eth.Topic("0x000000000000000000000000896dd350806eba53dfa9778c4698224e8ede2c41")},
				},
			},
		},
		{
			Message: "extraneous null topics should be filtered",
			Payload: `{"topics": [null]}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "extraneous null topics should be filtered",
			Payload: `{"topics": ["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", null]}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics: [][]eth.Topic{
					{eth.Topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
				},
			},
		},
		{
			Message: "null topics should be supported",
			Payload: `{"topics": [null, "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"]}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics: [][]eth.Topic{
					{},
					{eth.Topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
				},
			},
		},
		{
			Message: "ORed topics should be supported",
			Payload: `{"topics": [["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", "0x9577941d28fff863bfbee4694a6a4a56fb09e169619189d2eaa750b5b4819995"]]}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics: [][]eth.Topic{
					{
						eth.Topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
						eth.Topic("0x9577941d28fff863bfbee4694a6a4a56fb09e169619189d2eaa750b5b4819995"),
					},
				},
			},
		},
		{
			Message: "blockHash must be supported",
			Payload: `{"blockHash":"0xb509a2149556380fbff167f2fdfad07cf9cfe8eb605e83298683008f46f419b5"}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: eth.MustHash("0xb509a2149556380fbff167f2fdfad07cf9cfe8eb605e83298683008f46f419b5"),
				Topics:    nil,
			},
		},
		{
			Message: "fromBlock must be supported",
			Payload: `{"fromBlock":"0x1234"}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: eth.MustBlockNumberOrTag("0x1234"),
				ToBlock:   nil,
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "toBlock must be supported",
			Payload: `{"toBlock":"0x1234"}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   eth.MustBlockNumberOrTag("0x1234"),
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "block tags must be supported, tags matching defaults should still be present",
			Payload: `{"fromBlock":"latest", "toBlock": "latest"}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: eth.MustBlockNumberOrTag("latest"),
				ToBlock:   eth.MustBlockNumberOrTag("latest"),
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "block tags must be supported",
			Payload: `{"fromBlock":"earliest", "toBlock": "earliest"}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: eth.MustBlockNumberOrTag("earliest"),
				ToBlock:   eth.MustBlockNumberOrTag("earliest"),
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "block tags must be supported",
			Payload: `{"fromBlock":"pending", "toBlock": "pending"}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: eth.MustBlockNumberOrTag(eth.TagPending.String()),
				ToBlock:   eth.MustBlockNumberOrTag(eth.TagPending.String()),
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "complex example",
			Payload: `{
				"fromBlock":"0x1234", 
				"toBlock": "latest", 
				"blockHash":"0xb509a2149556380fbff167f2fdfad07cf9cfe8eb605e83298683008f46f419b5", 
				"address": ["0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c", "0x5717adf502fd8830456bd5dc26801a4db394e6b2"],
				"topics": [["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", "0x9577941d28fff863bfbee4694a6a4a56fb09e169619189d2eaa750b5b4819995"], "0x000000000000000000000000896dd350806eba53dfa9778c4698224e8ede2c41"]
			}`,
			Expected: eth.LogFilter{
				Address: []eth.Address{
					*eth.MustAddress("0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c"),
					*eth.MustAddress("0x5717adf502fd8830456bd5dc26801a4db394e6b2"),
				},
				FromBlock: eth.MustBlockNumberOrTag("0x1234"),
				ToBlock:   eth.MustBlockNumberOrTag("latest"),
				BlockHash: eth.MustHash("0xb509a2149556380fbff167f2fdfad07cf9cfe8eb605e83298683008f46f419b5"),
				Topics: [][]eth.Topic{
					{
						eth.Topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
						eth.Topic("0x9577941d28fff863bfbee4694a6a4a56fb09e169619189d2eaa750b5b4819995"),
					},
					{eth.Topic("0x000000000000000000000000896dd350806eba53dfa9778c4698224e8ede2c41")},
				},
			},
		},
		{
			Message:  "invalid topics should fail",
			Payload:  `{"topics":["0xfoo4foo4foo4foo4foo4foo4foo4foo4foo4foo4foo4foo4foo4foo4foo4foo4"]}`,
			Expected: eth.LogFilter{},
			Error:    true,
		},
		{
			Message:  "invalid topic array should fail",
			Payload:  `{"topics":[["0xfoo4foo4foo4foo4foo4foo4foo4foo4foo4foo4foo4foo4foo4foo4foo4foo4"]]}`,
			Expected: eth.LogFilter{},
			Error:    true,
		},
	}

	for _, test := range tests {
		actual := eth.LogFilter{}
		err := json.Unmarshal([]byte(test.Payload), &actual)
		if test.Error {
			require.Error(t, err, test.Message)
		} else {
			require.NoError(t, err, test.Message)
			require.Equal(t, test.Expected, actual, test.Message)
		}
	}
}
