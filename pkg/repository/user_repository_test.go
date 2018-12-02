package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/mimir-news/pkg/schema/stock"
	"github.com/mimir-news/pkg/schema/user"
	"github.com/stretchr/testify/assert"
)

var (
	testNow          = time.Now()
	test1HourFromNow = time.Now().Add(1 * time.Hour)
	test2HourFromNow = time.Now().Add(2 * time.Hour)
	test3HourFromNow = time.Now().Add(3 * time.Hour)
)

var testMembers = []watchlistMember{
	watchlistMember{
		listID:        "0",
		listName:      "l-0",
		listCreatedAt: testNow,
		stockSymbol:   sql.NullString{String: "S0", Valid: true},
		stockName:     sql.NullString{String: "s-0", Valid: true},
	},
	watchlistMember{
		listID:        "2",
		listName:      "l-2",
		listCreatedAt: test2HourFromNow,
		stockSymbol:   sql.NullString{String: "S5", Valid: true},
		stockName:     sql.NullString{String: "s-5", Valid: true},
	},
	watchlistMember{
		listID:        "1",
		listName:      "l-1",
		listCreatedAt: test1HourFromNow,
		stockSymbol:   sql.NullString{String: "S7", Valid: true},
		stockName:     sql.NullString{String: "s-7", Valid: true},
	},
	watchlistMember{
		listID:        "1",
		listName:      "l-1",
		listCreatedAt: test1HourFromNow,
		stockSymbol:   sql.NullString{String: "S3", Valid: true},
		stockName:     sql.NullString{String: "s-3", Valid: true},
	},
	watchlistMember{
		listID:        "2",
		listName:      "l-2",
		listCreatedAt: test2HourFromNow,
		stockSymbol:   sql.NullString{String: "S1", Valid: true},
		stockName:     sql.NullString{String: "s-1", Valid: true},
	},
	watchlistMember{
		listID:        "0",
		listName:      "l-0",
		listCreatedAt: testNow,
		stockSymbol:   sql.NullString{String: "S2", Valid: true},
		stockName:     sql.NullString{String: "s-2", Valid: true},
	},
	watchlistMember{
		listID:        "0",
		listName:      "l-0",
		listCreatedAt: testNow,
		stockSymbol:   sql.NullString{Valid: false},
		stockName:     sql.NullString{Valid: false},
	},
	watchlistMember{
		listID:        "3",
		listName:      "l-3",
		listCreatedAt: test3HourFromNow,
		stockSymbol:   sql.NullString{Valid: false},
		stockName:     sql.NullString{Valid: false},
	},
	watchlistMember{
		listID:        "3",
		listName:      "l-3",
		listCreatedAt: test3HourFromNow,
		stockSymbol:   sql.NullString{Valid: false},
		stockName:     sql.NullString{Valid: false},
	},
}

var expectedWatchlists = []user.Watchlist{
	user.Watchlist{
		ID:        "0",
		Name:      "l-0",
		CreatedAt: testNow,
		Stocks: []stock.Stock{
			stock.Stock{Symbol: "S0", Name: "s-0"},
			stock.Stock{Symbol: "S2", Name: "s-2"},
		},
	},
	user.Watchlist{
		ID:        "1",
		Name:      "l-1",
		CreatedAt: test1HourFromNow,
		Stocks: []stock.Stock{
			stock.Stock{Symbol: "S7", Name: "s-7"},
			stock.Stock{Symbol: "S3", Name: "s-3"},
		},
	},
	user.Watchlist{
		ID:        "2",
		Name:      "l-2",
		CreatedAt: test2HourFromNow,
		Stocks: []stock.Stock{
			stock.Stock{Symbol: "S5", Name: "s-5"},
			stock.Stock{Symbol: "S1", Name: "s-1"},
		},
	},
	user.Watchlist{
		ID:        "3",
		Name:      "l-3",
		CreatedAt: test3HourFromNow,
		Stocks:    []stock.Stock{},
	},
}

func TestCreateWatchlists(t *testing.T) {
	assert := assert.New(t)

	actualLists := createWatchlists(testMembers)
	assert.Equal(len(expectedWatchlists), len(actualLists))

	for i, watchlist := range actualLists {
		el := expectedWatchlists[i]
		assert.Equal(el.ID, watchlist.ID)
		assert.Equal(el.Name, watchlist.Name)
		assert.Equal(el.CreatedAt, watchlist.CreatedAt)
		assert.Equal(len(el.Stocks), len(el.Stocks))

		for j, s := range watchlist.Stocks {
			es := el.Stocks[j]
			assert.Equal(es.Name, s.Name)
			assert.Equal(es.Symbol, s.Symbol)
		}
	}
}
