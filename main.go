package main

import (
  "fmt"
  "github.com/julienschmidt/httprouter"
  "net/http"
  "encoding/json"
  "log"
  "github.com/levi/capycoin/blockchain"
  "github.com/levi/capycoin/hashcash"
  "github.com/levi/capycoin/nodes"
  "github.com/satori/go.uuid"
  "io/ioutil"
  "strings"
)

type ErrorResponse struct {
  err string `json:"error"`
}

type ChainResponse struct {
  Chain []blockchain.Block `json:"chain"`
  Length int `json:"length"`
}

type TransactionRequest struct {
  Sender string `json:"sender"`
  Recipient string `json:"recipient"`
  Amount int `json:"amount"`
}

type MineResponse struct {
  Message string `json:"message"`
  Index int `json:"index"`
  Transactions []blockchain.Transaction `json:"transactions"`
  Proof int `json:"proof"`
  PrevHash string `json:"prev_hash"`
}

type RegisterNodeRequest struct {
  Nodes map[string]string `json:"nodes"`
}

type RegisterNodeResponse struct {
  Message string `json:"message"`
  Nodes map[string]string `json:"nodes"`
}

type ResolveNodeResponse struct {
  Message string `json:"message"`
  Chain []blockchain.Block `json:"chain"`
}

func JSONResponse(handler http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("Content-Type", "application/json")
    handler.ServeHTTP(w, r)
  })
}

func main() {
  router := httprouter.New()

  u4 := uuid.Must(uuid.NewV4())
  nodeId := strings.Replace(u4.String(), "-", "", -1)
  bc := blockchain.New()

  nodes := nodes.New()

  router.GET("/", func (w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprint(w, "{ \"status\": \"ok\" }")
  })

  router.GET("/chain", func (w http.ResponseWriter, r *http.Request, _ httprouter.Params)  {
    j, err := json.Marshal(bc)

    if err != nil {
      resp := ErrorResponse{"Encoding JSON failed"}
      r, _ := json.Marshal(resp)
      w.Write(r)
      return
    }

    c := ChainResponse{
      bc.Chain,
      len(bc.Chain),
    }
    j, err = json.Marshal(c)
    if err != nil {
      // TODO: Handle error
      fmt.Printf("Error marhsalling json: %v\n", err)
      return
    }
    w.Write(j)
  })

  router.POST("/transactions/new", func (w http.ResponseWriter, r *http.Request, _ httprouter.Params)  {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
      // TODO: Handle error
      fmt.Printf("Error reading request body: %v\n", err)
      return
    }

    var p TransactionRequest
    err = json.Unmarshal(body, &p)
    if err != nil {
      // TODO: Handle error
      fmt.Printf("Error marhsalling json: %v\n", err)
      return
    }
    index := bc.NewTransaction(p.Sender, p.Recipient, p.Amount)

    fmt.Fprintf(w, "{ \"message\": \"Transaction will be added to Block %d\" }", index)
  })

  router.GET("/mine", func (w http.ResponseWriter, r *http.Request, _ httprouter.Params)  {
    lastBlock := bc.LastBlock()
    lastProof := lastBlock.Proof
    proof := hashcash.ProofOfWork(lastProof)

    bc.NewTransaction("0", nodeId, 1)

    prevHash, err := lastBlock.Hash()
    if err != nil {
      // TODO: Handle error
      fmt.Printf("Error producing block hash: %v\n", err)
      return
    }

    block := bc.NewBlock(proof, prevHash)

    m := MineResponse{
      "New Block Forged",
      block.Index,
      block.Transactions,
      block.Proof,
      block.PrevHash,
    }
    j, err := json.Marshal(m)
    if err != nil {
      // TODO: Handle error
      fmt.Printf("Error marhsalling json: %v\n", err)
      return
    }
    w.Write(j)
  })

  router.POST("/nodes/register", func (w http.ResponseWriter, r *http.Request, _ httprouter.Params)  {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
      // TODO: Handle error
      fmt.Printf("Error reading request body: %v\n", err)
      return
    }

    var n RegisterNodeRequest
    err = json.Unmarshal(body, &n)
    if err != nil {
      // TODO: Handle error
      fmt.Printf("Error marhsalling json: %v\n", err)
      return
    }

    for url, address := range n.Nodes {
      nodes.Register(url, address)
    }

    m := RegisterNodeResponse{
      "New nodes have been added",
      nodes.Addresses,
    }
    j, err := json.Marshal(m)
    if err != nil {
      // TODO: Handle error
      fmt.Printf("Error marhsalling json: %v\n", err)
      return
    }
    w.Write(j)
  })

  router.POST("/nodes/resolve", func (w http.ResponseWriter, r *http.Request, _ httprouter.Params)  {
    replaced, err := bc.ResolveConflicts(nodes)
    if err != nil {
      // TODO: Handle error
      fmt.Printf("Unable to resolve conflicts: %v\n", err)
      return
    }
    var m ResolveNodeResponse
    if replaced {
      m = ResolveNodeResponse{
        "Our chain was replaced",
        bc.Chain,
      }
    } else {
      m = ResolveNodeResponse{
        "Out chain was the authority",
        bc.Chain,
      }
    }
    j, err := json.Marshal(m)
    if err != nil {
      // TODO: Handle error
      fmt.Printf("Error marhsalling json: %v\n", err)
      return
    }
    w.Write(j)
  })

  fmt.Printf("Listening on localhost:3000\n")
  err := http.ListenAndServe(":3000", JSONResponse(router))
  log.Fatal(err)
}
