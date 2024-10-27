package interceptor

import (
	"embed"
	"github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/22 22:35
 * @file: i18n_interceptor.go
 * @description: i18n 国际化拦截器
 */

//go:embed ../../../conf.d/i18n/*
var i18nFS embed.FS

func I18nInterceptor() gin.HandlerFunc {
	return i18n.Localize(
		i18n.WithBundle(&i18n.BundleCfg{
			DefaultLanguage:  language.English,
			RootPath:         "../../../conf.d/i18n",
			AcceptLanguage:   []language.Tag{language.Chinese, language.English},
			FormatBundleFile: "yaml",
			UnmarshalFunc:    yaml.Unmarshal,
			Loader: &i18n.EmbedLoader{
				FS: i18nFS,
			},
		}),
		i18n.WithGetLngHandle(
			func(context *gin.Context, defaultLng string) string {
				lng := context.GetHeader("Accept-Language")
				if lng == "" {
					return defaultLng
				}
				return lng
			},
		),
	)
}
