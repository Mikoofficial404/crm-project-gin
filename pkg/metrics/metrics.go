package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crm_http_requests_total",
			Help: "Total jumlah HTTP request masuk",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "crm_http_request_duration_seconds",
			Help:    "Durasi HTTP request dalam detik",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	ActiveLeads = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "crm_active_leads_total",
			Help: "Total lead yang aktif (belum closed)",
		},
	)

	ActiveDeals = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "crm_active_deals_total",
			Help: "Total deal yang aktif",
		},
	)

	WhatsAppMessagesIn = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "crm_whatsapp_messages_in_total",
			Help: "Total pesan WhatsApp masuk",
		},
	)

	WhatsAppMessagesOut = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "crm_whatsapp_messages_out_total",
			Help: "Total pesan WhatsApp yang dikirim",
		},
	)
)

func Init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		ActiveLeads,
		ActiveDeals,
		WhatsAppMessagesIn,
		WhatsAppMessagesOut,
	)
}
