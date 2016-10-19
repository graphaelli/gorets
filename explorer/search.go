package explorer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jpfielding/gorets/rets"
)

// SearchParams ...
type SearchParams struct {
	Resource  string `json:"resource"`
	Class     string `json:"class"`
	Format    string `json:"format"`
	QueryType string `json:"query-type"`
	CountType int    `json:"count-type"`
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
	Query     string `json:"query"`
	Select    string `json:"select"`
}

// SearchPage ...
type SearchPage struct {
	Columns rets.Row   `json:"columns"`
	Rows    []rets.Row `json:"rows"`
}

// Search ...
// input: SearchOptions
// output: {
//   "columns": ["a","b","c"],
//   "rows": [
//       ["1","11","111"],
//       ["2","22","222"],
//       ["3","33","333"],
//   ]
// }
func Search(ctx context.Context, c *Connection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var p SearchParams
		if r.Body == nil {
			http.Error(w, "missing search params", 400)
			return
		}
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		fmt.Printf("params: %v\n", p)
		if p.QueryType == "" {
			p.QueryType = "DQML2"
		}
		if p.Format == "" {
			p.Format = "COMPACT_DECODED"
		}

		req := rets.SearchRequest{
			URL: c.URLs.Search,
			SearchParams: rets.SearchParams{
				Select:     p.Select,
				Query:      p.Query,
				SearchType: p.Resource,
				Class:      p.Class,
				Format:     p.Format,
				QueryType:  p.QueryType,
				Count:      p.CountType,
				Limit:      p.Limit,
				Offset:     p.Offset,
			},
		}

		fmt.Printf("Querying : %v\n", req)
		result, err := rets.SearchCompact(c.Requester, ctx, req)
		defer result.Close()
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		// opening the strea
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		w.Write([]byte("{"))

		w.Write([]byte("\n\"columns\": "))
		enc.Encode(result.Columns)
		w.Write([]byte("\n,\"rows\": ["))
		comma := false
		// too late to err in http here, need another solution
		result.ForEach(func(row rets.Row, err error) error {
			if comma {
				w.Write([]byte(","))
			}
			enc.Encode(row)
			comma = true
			return nil
		})
		w.Write([]byte("]"))
		w.Write([]byte("}"))
	}
}