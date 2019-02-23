package redirect

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/tschokko/mdthklb/config"
	"github.com/tschokko/mdthklb/pkg/roundrobin"
)

type Handler struct {
	wrr *roundrobin.WeightedRoundRobin
}

func NewHandler(servers []config.Server) *Handler {
	wrr := roundrobin.NewWeightedRoundRobin()
	for _, srv := range servers {
		wrr.AppendServer(srv.URL, roundrobin.Weight(srv.Weight))
	}

	return &Handler{wrr}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.Match([]string{http.MethodGet, http.MethodHead}, "/*", h.redirectHandler)
}

func (h *Handler) redirectHandler(c echo.Context) error {
	url, err := h.wrr.Next()
	if err != nil {
		// TODO: a panic isn't very cool!
		panic(err)
	}

	return c.Redirect(http.StatusTemporaryRedirect,
		fmt.Sprintf("%s%s", url, c.Request().RequestURI))
}
