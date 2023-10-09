package main

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2/klogr"
	"log"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func main() {
	var logger = klogr.New()
	// 初始化admission decoder
	scheme := runtime.NewScheme()
	decoder, err := admission.NewDecoder(scheme)
	if err != nil {
		log.Fatalln(err)
	}

	// 初始化MutatingHandler
	mutation := MutatingHandler{
		Decoder: decoder,
	}
	webhook := admission.Webhook{
		Handler: &mutation,
	}

	_, err = inject.LoggerInto(logger, &webhook)
	if err != nil {
		log.Fatalln(err)
	}
	// 注册mutation-pod 回调接口
	http.HandleFunc("/mutation-pod", webhook.ServeHTTP)
	// TLS证书和密钥配置
	err = http.ListenAndServeTLS(":8000", "tls/webhook.pem", "tls/webhook-key.pem", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

