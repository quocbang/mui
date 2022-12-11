package middleware

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/xid"
	"go.uber.org/zap"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"

	handlerUtils "gitlab.kenda.com.tw/kenda/mui/server/impl/handlers/mcom/utils"
)

// LoggingMiddleware adds request ids and logs responses.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := xid.New().String()
		logger := zap.L().With(zap.String("rid", requestID))
		r = r.WithContext(commonsCtx.WithLogger(r.Context(), logger))
		r = handlerUtils.SetContextValue(r, "rid", requestID)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		var respBuf bytes.Buffer
		ww.Tee(&respBuf)

		defer logger.Sync() // nolint

		if r.Method == "" {
			r.Method = http.MethodGet
		}

		reqLogFields := []zap.Field{}
		// 登入畫面不需要印出請求資訊
		if !strings.Contains(r.URL.Path, "user/login") &&
			!strings.Contains(r.URL.Path, "user/change-password") {
			request, err := httputil.DumpRequest(r, r.Method != http.MethodGet)
			reqLogFields = append(reqLogFields,
				zap.String("request", string(request)),
				zap.NamedError("request_dump_error", err),
			)
		}
		logger.Info("start request", reqLogFields...)

		defer func(start time.Time) {
			responseFields := []zap.Field{
				zap.String("request_method", r.Method),
				zap.String("request_url", r.URL.Path),
				zap.Int("status_code", ww.Status()),
				zap.String("remote_address", r.RemoteAddr),
				zap.String("x-forwarded-for", r.Header.Get("x-forwarded-for")),
				zap.String("duration", time.Since(start).String()),
			}
			// "application/pdf" as response header's content-type do not print response data
			contentTypes := ww.Header().Values("Content-Type")
			for _, contentType := range contentTypes {
				if contentType == "application/json" {
					responseFields = append(responseFields, zap.String("response", respBuf.String()))
					break
				}
			}
			logger.Info("server responses", responseFields...)
		}(time.Now())

		next.ServeHTTP(ww, r)
	})
}
