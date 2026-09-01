package api

import (
	"net/http"
	"strings"

	"github.com/forgeflow/forgeflow/internal/auth"
)

type authHandler struct{service *auth.Service}
func(h authHandler)register(w http.ResponseWriter,r *http.Request){var input auth.RegisterInput;if err:=decodeJSON(w,r,&input);err!=nil{writeError(w,r,err);return};user,err:=h.service.Register(r.Context(),input);if err!=nil{writeError(w,r,err);return};user.PasswordHash="";writeJSON(w,http.StatusCreated,user)}
func(h authHandler)login(w http.ResponseWriter,r *http.Request){var input struct{Email string `json:"email"`;Password string `json:"password"`};if err:=decodeJSON(w,r,&input);err!=nil{writeError(w,r,err);return};session,err:=h.service.Login(r.Context(),input.Email,input.Password);if err!=nil{writeError(w,r,err);return};writeJSON(w,http.StatusOK,session)}
func(h authHandler)logout(w http.ResponseWriter,r *http.Request){token:=strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"),"Bearer "));if err:=h.service.Logout(r.Context(),token);err!=nil{writeError(w,r,err);return};w.WriteHeader(http.StatusNoContent)}
func(h authHandler)me(w http.ResponseWriter,r *http.Request){writeJSON(w,http.StatusOK,currentUser(r.Context()))}

