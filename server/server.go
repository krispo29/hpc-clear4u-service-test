package server

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/go-pg/pg/v9"

	"hpc-express-service/constant"
	"hpc-express-service/factory"
)

type Server struct {
	router         chi.Router
	svcFactory     *factory.ServiceFactory
	postgreSQLConn *pg.DB
	mode           string
}

var (
	TokenAuthRSA256 *jwtauth.JWTAuth
	ResponseSuccess = "success"
	ResponseFailed  = "failed"
)
var priBold, priRegular, priLight, frontTHSarabunNew, frontTHSarabunNewBold, frontTHSarabunNewBoldItalic, frontTHSarabunNewItalic []byte

func init() {
	var err error

	// Loading auth
	key, err := ioutil.ReadFile("public.pem")
	if err != nil {
		log.Fatal(err)
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(key)
	if err != nil {
		log.Fatal(err)
	}
	TokenAuthRSA256 = jwtauth.New("RS256", nil, publicKey)
	// Loading Font
	frontTHSarabunNew, err = ioutil.ReadFile("assets/THSarabunNew.ttf")
	if err != nil {
		log.Println(err)
	}

	frontTHSarabunNewBold, err = ioutil.ReadFile("assets/THSarabunNew Bold.ttf")
	if err != nil {
		log.Println(err)
	}

	frontTHSarabunNewBoldItalic, err = ioutil.ReadFile("assets/THSarabunNew BoldItalic.ttf")
	if err != nil {
		log.Println(err)
	}

	frontTHSarabunNewItalic, err = ioutil.ReadFile("assets/THSarabunNew Italic.ttf")
	if err != nil {
		log.Println(err)
	}

}

func New(
	svcFactory *factory.ServiceFactory,
	postgreSQLConn *pg.DB,
	mode string,
) *Server {
	s := &Server{
		svcFactory:     svcFactory,
		postgreSQLConn: postgreSQLConn,
		mode:           mode,
	}
	r := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "*.web.app", "https://app-staging.clear4u.co", "https://app.clear4u.co"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(addCtx("postgreSQLConn", s.postgreSQLConn))
	r.Use(addCtx("mode", s.mode))

	// Create a route along /static that will serve contents from
	// the ./uploads/ folder.
	// workDir, _ := os.Getwd()
	// filesDir := http.Dir(filepath.Join(workDir, "uploads"))
	// FileServer(r, "/static", filesDir)

	r.Group(func(r chi.Router) {

		r.Route("/v1", func(r chi.Router) {

			r.Use(jwtauth.Verifier(TokenAuthRSA256))
			r.Use(jwtauth.Authenticator(TokenAuthRSA256))

			dashboardSvc := dashboardHandler{s.svcFactory.DashboardSvc}
			r.Mount("/dashboard", dashboardSvc.router())

			commonSvc := commonHandler{s.svcFactory.CommonSvc}
			r.Mount("/common", commonSvc.router())

			customerSvc := customerHandler{s.svcFactory.CustomerSvc}
			r.Mount("/customers", customerSvc.router())

			userSvc := userHandler{s.svcFactory.UserSvc}
			r.Mount("/users", userSvc.router())

			uploadlogSvc := uploadLoggingHandler{s.svcFactory.UploadlogSvc}
			r.Mount("/uploadlog", uploadlogSvc.router())

			mawbSvc := mawbHandler{s.svcFactory.MawbSvc}
			r.Mount("/mawb", mawbSvc.router())

			settingSvc := settingHandler{s.svcFactory.SettingSvc}
			r.Mount("/settings", settingSvc.router())

			dropdownSvc := dropdownHandler{s.svcFactory.DropdownSvc}
			r.Mount("/dropdown", dropdownSvc.router())

			mawbInfoSvc := mawbInfoHandler{
				s:                s.svcFactory.MawbInfoSvc,
				cargoManifestSvc: s.svcFactory.CargoManifestSvc,
				draftMAWBSvc:     s.svcFactory.DraftMAWBSvc,
			}
			r.Mount("/mawbinfo", mawbInfoSvc.router())

			compareSvc := excelHandler{s.svcFactory.CompareSvc}
			r.Mount("/compare", compareSvc.router())

			//("/compare", compareApiHandler.CompareExcel)

			// r.Mount("/inbound", manifestSvc.router())

			r.Route("/inbound", func(r chi.Router) {
				inboundExpressServiceSvc := inboundExpressHandler{s.svcFactory.InboundExpressServiceSvc}
				r.Mount("/express", inboundExpressServiceSvc.router())
			})

			r.Route("/outbound", func(r chi.Router) {
				outboundExpressServiceSvc := outboundExpressHandler{s.svcFactory.OutboundExpressServiceSvc}
				r.Mount("/express", outboundExpressServiceSvc.router())

				outboundMawbServiceSvc := outboundMawbHandler{s.svcFactory.OutboundMawbServiceSvc}
				r.Mount("/mawb", outboundMawbServiceSvc.router())

			})

			// authSvc := authHandler{s.svcFactory.AuthSvc}
			// r.Mount("/manifest", authSvc.router())

		})
	})

	// Public
	r.Group(func(r chi.Router) {
		r.Route("/", func(r chi.Router) {
			authSvc := authHandler{s.svcFactory.AuthSvc}
			r.Mount("/auth", authSvc.router())
		})

	})

	// Protected
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(TokenAuthRSA256))
		r.Use(jwtauth.Authenticator(TokenAuthRSA256))

		r.Get("/signed", func(w http.ResponseWriter, r *http.Request) {
			render.Respond(w, r, SuccessResponse(nil, "OK"))
		})
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		render.Respond(w, r, SuccessResponse(nil, "OK"))
	})

	s.router = r

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		route = strings.Replace(route, "/*/", "/", -1)
		log.Printf("%s %s\n", method, route) // Walk and print out all routes
		return nil
	}

	if err := chi.Walk(r, walkFunc); err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func addCtx(name string, db interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), name, db))
			next.ServeHTTP(w, r)
		})
	}
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status,omitempty"`  // user-level status message
	AppCode    int64  `json:"code,omitempty"`    // application-specific error code
	Message    string `json:"message,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		AppCode:        constant.CodeError,
		// StatusText:     "Invalid request.",
		Message: err.Error(),
	}
}

func ErrUnauthorized(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusUnauthorized,
		AppCode:        constant.CodeUnauthorized,
		// StatusText:     "Invalid request.",
		Message: err.Error(),
	}
}

type ApiResponse struct {
	HTTPStatusCode int `json:"-"` // http response status code

	AppCode int64       `json:"code,omitempty"` // application-specific error code
	Data    interface{} `json:"data,omitempty"` // application-specific error code
	Message string      `json:"message"`        // application-level error message, for debugging
}

func (e *ApiResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func SuccessResponse(data interface{}, msg string) render.Renderer {
	return &ApiResponse{
		HTTPStatusCode: http.StatusOK,
		AppCode:        constant.CodeSuccess,
		Data:           data,
		Message:        msg,
	}
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func GetUserUUIDFromContext(r *http.Request) string {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return ""
	}

	return claims["uuid"].(string)
}
