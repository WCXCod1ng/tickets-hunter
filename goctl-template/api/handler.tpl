// Code scaffolded by goctl. Safe to edit.
// goctl {{.version}}

package {{.PkgName}}

import (
    "net/http"
    "tickets-hunter/common/result" // 添加这一行引用
    "github.com/zeromicro/go-zero/rest/httpx"
    {{.ImportPackages}}
)

func {{.HandlerName}}(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        {{if .HasRequest}}var req types.{{.RequestType}}
        if err := httpx.Parse(r, &req); err != nil {
            result.HttpResult(r, w, nil, err) // 解析失败也走统一处理
            return
        }

        {{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), svcCtx)
        {{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}&req{{end}})

        // 使用自定义的 HttpResult
        result.HttpResult(r, w, {{if .HasResp}}resp{{else}}nil{{end}}, err)
    }
}