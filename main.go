package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"opspilot/internal/auth"
	"opspilot/internal/config"
	"opspilot/internal/crypto"
	"opspilot/internal/metrics"
	"opspilot/internal/models"
	"opspilot/internal/visualizer"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moby/moby/client"
)

func main() {
	// 1. Initialize Database
	db := config.InitDB(nil)

	// 2. Initialize Docker Client
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Warning: Failed to initialize Docker client: %v", err)
	}

	// 3. Initialize Metrics
	collector := &metrics.MetricCollector{
		Docker:             dockerCli,
		VictoriaMetricsURL: os.Getenv("VICTORIAMETRICS_URL"),
	}
	streamer := &metrics.MetricStreamer{
		Collector: collector,
	}

	// 4. Initialize Visualizer
	viz := visualizer.NewOpsVisualizer(db)

	// 5. Initialize Router
	r := gin.Default()

	// Load templates
	r.LoadHTMLGlob("ui/templates/*")
	r.Static("/static", "./ui/static")

	// 6. Routes
	r.GET("/", func(c *gin.Context) {
		var envCount int64
		db.Model(&models.Environment{}).Count(&envCount)

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title":     "Dashboard",
			"EnvsCount": envCount,
		})
	})

	// MFA Enrollment
	r.GET("/auth/mfa/enroll", func(c *gin.Context) {
		qrBase64, secret, err := auth.GenerateTOTPQRCode("admin@opspilot.local")
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Failed to generate MFA secret"})
			return
		}

		c.HTML(http.StatusOK, "mfa_enroll.html", gin.H{
			"QRCode": template.URL(fmt.Sprintf("data:image/png;base64,%s", qrBase64)),
			"Secret": secret,
		})
	})

	// Audit Logs Page
	r.GET("/audit", func(c *gin.Context) {
		c.HTML(http.StatusOK, "audit_viewer.html", nil)
	})

	// Target Hosts Page
	r.GET("/hosts", func(c *gin.Context) {
		c.HTML(http.StatusOK, "hosts.html", nil)
	})

	// Hosts API (HTMX)
	r.GET("/api/hosts", func(c *gin.Context) {
		var hosts []models.TargetHost
		db.Order("name asc").Find(&hosts)

		html := ""
		for _, host := range hosts {
			html += fmt.Sprintf(`
				<tr class="hover:bg-slate-50 transition-colors">
					<td class="px-6 py-4 font-medium text-slate-800">%s</td>
					<td class="px-6 py-4"><span class="px-2 py-1 bg-slate-100 text-slate-600 rounded text-xs font-bold uppercase">%s</span></td>
					<td class="px-6 py-4 text-slate-500 font-mono text-xs">%s</td>
					<td class="px-6 py-4">
						<button class="text-indigo-600 hover:text-indigo-900 font-bold">Edit</button>
					</td>
				</tr>`,
				host.Name,
				host.Type,
				host.Endpoint,
			)
		}
		if html == "" {
			html = "<tr><td colspan='4' class='px-6 py-8 text-center text-slate-400'>No target hosts registered.</td></tr>"
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	r.POST("/api/hosts", func(c *gin.Context) {
		name := c.PostForm("name")
		hostType := c.PostForm("type")
		endpoint := c.PostForm("endpoint")
		authData := c.PostForm("auth_data")

		// Encrypt sensitive data if provided
		encryptedAuth := ""
		if authData != "" {
			var err error
			encryptedAuth, err = crypto.Encrypt(authData)
			if err != nil {
				c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte("<tr><td colspan='4' class='text-red-500 p-4'>Encryption failed</td></tr>"))
				return
			}
		}

		host := models.TargetHost{
			Name:     name,
			Type:     hostType,
			Endpoint: endpoint,
			AuthData: encryptedAuth,
		}

		if err := db.Create(&host).Error; err != nil {
			c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte("<tr><td colspan='4' class='text-red-500 p-4'>Database error</td></tr>"))
			return
		}

		// Return updated list
		c.Redirect(http.StatusSeeOther, "/api/hosts")
	})

	// Audit API (HTMX)
	r.GET("/api/audit", func(c *gin.Context) {
		var logs []models.AuditLog
		db.Order("created_at desc").Limit(50).Find(&logs)

		html := ""
		for _, log := range logs {
			html += fmt.Sprintf(`
				<tr class="hover:bg-slate-50 transition-colors">
					<td class="px-6 py-4 text-slate-600">%s</td>
					<td class="px-6 py-4 font-mono text-xs text-slate-500">%s</td>
					<td class="px-6 py-4"><span class="px-2 py-1 bg-indigo-100 text-indigo-700 rounded text-xs font-bold">%s</span></td>
					<td class="px-6 py-4 text-slate-800 font-medium">%s</td>
					<td class="px-6 py-4 text-slate-500">%s</td>
				</tr>`,
				log.CreatedAt.Format("2006-01-02 15:04:05"),
				log.UserID,
				log.Action,
				log.Target,
				log.IPAddress,
			)
		}
		if html == "" {
			html = "<tr><td colspan='5' class='px-6 py-8 text-center text-slate-400'>No audit logs found.</td></tr>"
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// Environment Wizard Page
	r.GET("/environments/new", func(c *gin.Context) {
		var hosts []models.TargetHost
		db.Order("name asc").Find(&hosts)

		c.HTML(http.StatusOK, "env_wizard.html", gin.H{
			"Hosts": hosts,
		})
	})

	// Environment API (HTMX Provisioning)
	r.POST("/api/environments", func(c *gin.Context) {
		name := c.PostForm("name")
		envType := c.PostForm("type")
		hostIDStr := c.PostForm("target_host_id")

		hostID, _ := uuid.Parse(hostIDStr)

		// Create record in DB
		env := models.Environment{
			Name:         name,
			Type:         envType,
			TargetHostID: &hostID,
			Status:       "PROVISIONING",
		}

		if err := db.Create(&env).Error; err != nil {
			c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(fmt.Sprintf(`
				<div class="mt-6 p-4 bg-red-100 border border-red-200 text-red-700 rounded-lg">
					<p class="font-bold">Error Initializing</p>
					<p class="text-sm">%s</p>
				</div>`, err.Error())))
			return
		}

		// Fetch host name for success message
		var host models.TargetHost
		db.First(&host, hostID)

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`
			<div class="mt-6 p-4 bg-green-100 border border-green-200 text-green-700 rounded-lg">
				<p class="font-bold">Success!</p>
				<p class="text-sm">Environment <strong>%s</strong> is now being provisioned on <strong>%s</strong>.</p>
				<a href="/" class="mt-2 inline-block text-xs font-bold uppercase tracking-wider underline">Go to Dashboard</a>
			</div>`, name, host.Name)))
	})

	// Topology API (HTMX support)
	r.GET("/api/topology", func(c *gin.Context) {
		nodes, edges := viz.BuildTopology()
		c.JSON(http.StatusOK, gin.H{
			"nodes": nodes,
			"edges": edges,
		})
	})

	// Topology WebSocket (Real-time updates)
	r.GET("/ws/topology", viz.StreamTopologyUpdates)

	// Live Metrics WebSocket
	r.GET("/ws/metrics/:id", streamer.StreamContainerStats)

	// Historical Metrics API
	r.GET("/api/metrics/history/:id", func(c *gin.Context) {
		id := c.Param("id")
		metric := c.Query("metric") // cpu_usage, memory_usage
		if metric == "" {
			metric = "cpu_usage"
		}

		query := fmt.Sprintf("docker_metrics_%s{container_id=\"%s\"}", metric, id)
		// Last 1 hour
		end := time.Now()
		start := end.Add(-1 * time.Hour)

		data, err := collector.QueryRange(c.Request.Context(), query, start, end, "1m")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Data(http.StatusOK, "application/json", []byte(data))
	})

	// 7. Start Server
	port := "8080"
	log.Printf("OpsPilot Control Plane starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
