package main

import (
    "net/http"
	"fmt"
    "github.com/gin-gonic/gin"
	"time"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// POST input expected infos
type product struct {
    Codigo string  `json:"id"`
    Nome  string  `json:"name"`
    EstoqueTotal int `json:"total_stock"`
    EstoqueDeCorte  int `json:"passing_stock"`
	PrecoDe float64 `json:"price_from"`
	PrecoPor float64 `json:"price_for"`
}
// DB stored infos
type stock_product struct {
	Codigo string  `json:"id"`
    Nome  string  `json:"name"`
    EstoqueTotal int `json:"total_stock"`
    EstoqueDeCorte  int `json:"passing_stock"`
	EstoqueDisponivel int `json:"available_stock"`
	PrecoDe float64 `json:"price_from"`
	PrecoPor float64 `json:"price_for"`
	UltimaModificacao time.Time `json:"last_change"`
}


func main() {
	// criando o bd
	db, err := gorm.Open(sqlite.Open("products.db"), &gorm.Config{})
    if err != nil {
        fmt.Print("Error: can't open db")
    }
	db.AutoMigrate(&stock_product{})
	handler := newHandler(db)
    router := gin.Default()
	router.LoadHTMLFiles("README.html")
    router.GET("/products", handler.getProducts)  // return all products
	router.GET("/product/:codigo", handler.getProductByID) // return 1 product
	router.POST("/product", handler.postProduct)  // include 1 product
	router.PUT("/product/:codigo", handler.updateProductById)  // update 1 product
	router.DELETE("/product/:codigo", handler.deleteProductByID)  // update 1 product
	router.GET("/", func(c *gin.Context){c.HTML(200, "README.html", gin.H{"message":"pass"})})
    router.Run("localhost:8080")
}

// store db handler
type Handler struct {
	db *gorm.DB
}

func newHandler(db *gorm.DB) *Handler {
	return &Handler{db}
}

// getAlbums responds with the list of all products as JSON.
func (h *Handler)getProducts(c *gin.Context) {
	var getStockProducts []stock_product
	if result := h.db.Find(&getStockProducts); result.Error != nil {
		c.IndentedJSON(http.StatusNotFound,gin.H{"message": "Nenhum produto encontrado"})
		return
	} 
    c.IndentedJSON(http.StatusOK, getStockProducts)
}

func (h *Handler) postProduct(c *gin.Context) {
    var newProduct product
	var newStockProduct stock_product

    // Call BindJSON to bind the received JSON to
    // new product
    if err := c.BindJSON(&newProduct); err != nil {  // lembra um for, mas sem o incremento
        fmt.Print(c)
		return
    }
	if newProduct.PrecoDe < newProduct.PrecoPor {
		fmt.Print("Preco POR maior que Preco DE")
		c.IndentedJSON(http.StatusNotFound,gin.H{"message": "Preco POR maior que Preco DE"})
		return
	}
	if err:=h.db.First(&newStockProduct,"Codigo = ?", newProduct.Codigo); err.Error==nil{
		c.IndentedJSON(http.StatusNotFound,gin.H{"message": "Já existe um produto com esse codigo, escolha outro"})
		return
	}
	
	// faz falta um dict de python
	newStockProduct.Codigo = newProduct.Codigo
    newStockProduct.Nome  = newProduct.Nome
    newStockProduct.EstoqueTotal = newProduct.EstoqueTotal
    newStockProduct.EstoqueDeCorte  = newProduct.EstoqueDeCorte
	newStockProduct.PrecoDe = newProduct.PrecoDe
	newStockProduct.PrecoPor = newProduct.PrecoPor
	newStockProduct.EstoqueDisponivel = newStockProduct.EstoqueTotal - newStockProduct.EstoqueDeCorte
	newStockProduct.UltimaModificacao = time.Now()
	h.db.Create(&newStockProduct)
    c.IndentedJSON(http.StatusCreated, newStockProduct)
}

func (h *Handler)getProductByID(c *gin.Context) {
    cod := c.Param("codigo")
	var out_product stock_product
	// find a product in db and render it
	if err:=h.db.First(&out_product,"Codigo = ?", cod); err.Error!=nil{
		c.IndentedJSON(http.StatusNotFound,gin.H{"message": "Produto nao encontrado"})
		return
	}
    c.IndentedJSON(http.StatusOK, out_product)
}

func (h *Handler)updateProductById(c *gin.Context) {
	var to_update product
	if err := c.BindJSON(&to_update); err != nil {  // lembra um for, mas sem o incremento
        fmt.Print(c) //nao conseguiu preencher os campos de produto
		return
    }
	cod := c.Param("codigo")
	var updt_stock_product stock_product
	if err:=h.db.First(&updt_stock_product,"Codigo = ?", cod); err.Error!=nil{
		c.IndentedJSON(http.StatusNotFound,gin.H{"message": "Produto nao encontrado /n addicione o produto com POST primeiro"})
		return
	}
	if to_update.PrecoDe < to_update.PrecoPor {
		fmt.Print("Preco POR maior que Preco DE")
		c.IndentedJSON(http.StatusNotFound,gin.H{"message": "Preco POR maior que Preco DE"})
		return
	}
	if err := h.db.First(&updt_stock_product,"Codigo = ?", to_update.Codigo); err.Error==nil{
		fmt.Print("Produto já existe escolha outro id")
		c.IndentedJSON(http.StatusNotFound,gin.H{"message":"Produto já existe escolha outro id"})
		return
	}
	h.db.First(&updt_stock_product,"Codigo = ?", cod)
	updt_stock_product.Codigo = to_update.Codigo
    updt_stock_product.Nome  = to_update.Nome
    updt_stock_product.EstoqueTotal = to_update.EstoqueTotal
    updt_stock_product.EstoqueDeCorte  = to_update.EstoqueDeCorte
	updt_stock_product.PrecoDe = to_update.PrecoDe
	updt_stock_product.PrecoPor = to_update.PrecoPor
	updt_stock_product.EstoqueDisponivel = updt_stock_product.EstoqueTotal - updt_stock_product.EstoqueDeCorte
	updt_stock_product.UltimaModificacao = time.Now()
	h.db.Debug().Where("Codigo = ?", cod).Save(&updt_stock_product)
	c.IndentedJSON(http.StatusOK, updt_stock_product)
	
}

func (h *Handler)deleteProductByID(c *gin.Context) {
    cod := c.Param("codigo")
	var del_product stock_product
	// find a product in db and render it
	if err:=h.db.First(&del_product,"Codigo = ?", cod); err.Error!=nil{
		c.IndentedJSON(http.StatusNotFound,gin.H{"message": "Produto nao encontrado"})
		return
	}
	h.db.First(&del_product,"Codigo = ?", cod)
	fmt.Print(del_product)
	h.db.Debug().Where("Codigo = ?",  del_product.Codigo).Delete(&del_product)
	if err:=h.db.First(&del_product,"Codigo = ?", cod); err.RowsAffected==0{
		c.IndentedJSON(http.StatusNotFound,gin.H{"message": "Produto deletado"})
		return
	}
    c.IndentedJSON(http.StatusOK,gin.H{"message": "Impossível deletar o produto"})
}
