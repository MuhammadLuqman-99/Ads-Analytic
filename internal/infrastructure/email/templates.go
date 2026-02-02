package email

// baseTemplate is the base HTML template for all emails
const baseTemplate = `
<!DOCTYPE html>
<html lang="ms">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Subject}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            margin: 0;
            padding: 0;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .email-wrapper {
            background-color: #ffffff;
            border-radius: 12px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #2563eb 0%, #7c3aed 100%);
            padding: 32px 24px;
            text-align: center;
        }
        .logo {
            color: #ffffff;
            font-size: 24px;
            font-weight: bold;
            text-decoration: none;
        }
        .content {
            padding: 32px 24px;
        }
        .footer {
            background-color: #f9fafb;
            padding: 24px;
            text-align: center;
            font-size: 12px;
            color: #6b7280;
        }
        h1 {
            color: #111827;
            font-size: 24px;
            margin: 0 0 16px 0;
        }
        p {
            color: #4b5563;
            margin: 0 0 16px 0;
        }
        .button {
            display: inline-block;
            background: linear-gradient(135deg, #2563eb 0%, #7c3aed 100%);
            color: #ffffff !important;
            text-decoration: none;
            padding: 14px 32px;
            border-radius: 8px;
            font-weight: 600;
            margin: 16px 0;
        }
        .button:hover {
            opacity: 0.9;
        }
        .metric-card {
            background-color: #f9fafb;
            border-radius: 8px;
            padding: 16px;
            margin: 8px 0;
            text-align: center;
        }
        .metric-value {
            font-size: 28px;
            font-weight: bold;
            color: #111827;
        }
        .metric-label {
            font-size: 12px;
            color: #6b7280;
            text-transform: uppercase;
        }
        .metric-change {
            font-size: 14px;
            margin-top: 4px;
        }
        .positive { color: #10b981; }
        .negative { color: #ef4444; }
        .warning-box {
            background-color: #fef3c7;
            border-left: 4px solid #f59e0b;
            padding: 16px;
            margin: 16px 0;
            border-radius: 0 8px 8px 0;
        }
        .success-box {
            background-color: #d1fae5;
            border-left: 4px solid #10b981;
            padding: 16px;
            margin: 16px 0;
            border-radius: 0 8px 8px 0;
        }
        .error-box {
            background-color: #fee2e2;
            border-left: 4px solid #ef4444;
            padding: 16px;
            margin: 16px 0;
            border-radius: 0 8px 8px 0;
        }
        .divider {
            border-top: 1px solid #e5e7eb;
            margin: 24px 0;
        }
        .social-links {
            margin-top: 16px;
        }
        .social-links a {
            display: inline-block;
            margin: 0 8px;
            color: #6b7280;
            text-decoration: none;
        }
        @media only screen and (max-width: 600px) {
            .container {
                padding: 10px;
            }
            .content {
                padding: 24px 16px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <a href="{{.BaseURL}}" class="logo">AdsAnalytic</a>
            </div>
            <div class="content">
                {{template "body" .}}
            </div>
            <div class="footer">
                <p>© 2026 AdsAnalytic Sdn Bhd. Hak cipta terpelihara.</p>
                <p>Anda menerima emel ini kerana anda berdaftar di AdsAnalytic.</p>
                <p>
                    <a href="{{.BaseURL}}/dashboard/settings/notifications">Urus notifikasi</a> |
                    <a href="{{.BaseURL}}/privacy">Polisi Privasi</a> |
                    <a href="{{.BaseURL}}/terms">Terma Perkhidmatan</a>
                </p>
                <div class="social-links">
                    <a href="https://facebook.com/adsanalytic">Facebook</a>
                    <a href="https://twitter.com/adsanalytic">Twitter</a>
                    <a href="https://instagram.com/adsanalytic">Instagram</a>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
`

// templates maps email types to their templates
var templates = map[EmailType]string{
	EmailTypeWelcome:             welcomeTemplate,
	EmailTypeVerification:        verificationTemplate,
	EmailTypePasswordReset:       passwordResetTemplate,
	EmailTypePlatformConnected:   platformConnectedTemplate,
	EmailTypeWeeklySummary:       weeklySummaryTemplate,
	EmailTypeTokenExpired:        tokenExpiredTemplate,
	EmailTypeSubscriptionConfirm: subscriptionConfirmTemplate,
	EmailTypePaymentFailed:       paymentFailedTemplate,
	EmailTypeConnectReminder:     connectReminderTemplate,
	EmailTypeInactiveReminder:    inactiveReminderTemplate,
}

// welcomeTemplate - sent after registration
const welcomeTemplate = `
{{define "body"}}
<h1>Selamat datang ke AdsAnalytic! 🎉</h1>

<p>Hai {{.Name}},</p>

<p>Terima kasih kerana mendaftar dengan AdsAnalytic! Kami teruja untuk membantu anda mengoptimumkan prestasi iklan e-commerce anda.</p>

<div class="success-box">
    <strong>Akaun anda sudah aktif!</strong><br>
    Anda boleh mula menggunakan AdsAnalytic sekarang.
</div>

<p>Langkah seterusnya:</p>
<ol>
    <li>Connect platform iklan pertama anda (Meta, TikTok, atau Shopee)</li>
    <li>Tunggu data sync (biasanya dalam 5 minit)</li>
    <li>Explore dashboard dan lihat insights anda!</li>
</ol>

<p style="text-align: center;">
    <a href="{{.BaseURL}}/dashboard/connections" class="button">Connect Platform Sekarang</a>
</p>

<div class="divider"></div>

<p><strong>Perlu bantuan?</strong></p>
<p>Tim support kami sedia membantu. Hubungi kami di <a href="mailto:support@adsanalytic.com">support@adsanalytic.com</a> atau layari <a href="{{.BaseURL}}/docs">dokumentasi</a> kami.</p>

<p>Selamat maju jaya!</p>
<p>— Tim AdsAnalytic</p>
{{end}}
`

// verificationTemplate - email verification
const verificationTemplate = `
{{define "body"}}
<h1>Sahkan alamat emel anda</h1>

<p>Hai {{.Name}},</p>

<p>Sila klik butang di bawah untuk mengesahkan alamat emel anda:</p>

<p style="text-align: center;">
    <a href="{{.VerificationURL}}" class="button">Sahkan Emel</a>
</p>

<p>Atau salin dan tampal URL ini ke pelayar anda:</p>
<p style="word-break: break-all; background-color: #f3f4f6; padding: 12px; border-radius: 6px; font-size: 14px;">
    {{.VerificationURL}}
</p>

<div class="warning-box">
    <strong>Pautan ini akan tamat tempoh dalam 24 jam.</strong>
</div>

<p>Jika anda tidak mendaftar akaun AdsAnalytic, sila abaikan emel ini.</p>

<p>— Tim AdsAnalytic</p>
{{end}}
`

// passwordResetTemplate - password reset
const passwordResetTemplate = `
{{define "body"}}
<h1>Reset kata laluan anda</h1>

<p>Hai {{.Name}},</p>

<p>Kami menerima permintaan untuk reset kata laluan akaun anda. Klik butang di bawah untuk menetapkan kata laluan baru:</p>

<p style="text-align: center;">
    <a href="{{.ResetURL}}" class="button">Reset Kata Laluan</a>
</p>

<div class="warning-box">
    <strong>Pautan ini akan tamat tempoh dalam 1 jam.</strong>
</div>

<p>Jika anda tidak meminta reset kata laluan, sila abaikan emel ini. Kata laluan anda tidak akan berubah.</p>

<p>Untuk keselamatan, permintaan ini datang dari:</p>
<ul>
    <li>IP Address: {{.IPAddress}}</li>
    <li>Masa: {{.RequestTime}}</li>
</ul>

<p>— Tim AdsAnalytic</p>
{{end}}
`

// platformConnectedTemplate - first platform connected congratulations
const platformConnectedTemplate = `
{{define "body"}}
<h1>Tahniah! 🎊 Platform berjaya disambungkan</h1>

<p>Hai {{.Name}},</p>

<p>Anda baru sahaja menyambungkan <strong>{{.PlatformName}}</strong> ke AdsAnalytic. Data iklan anda sedang di-sync sekarang!</p>

<div class="success-box">
    <strong>{{.PlatformName}} berjaya disambungkan!</strong><br>
    Data anda akan tersedia dalam dashboard dalam beberapa minit.
</div>

<p>Apa yang akan berlaku seterusnya:</p>
<ol>
    <li>✅ Data iklan akan di-sync secara automatik</li>
    <li>📊 Dashboard akan memaparkan metrik anda</li>
    <li>📈 Anda boleh mula track ROAS dan prestasi</li>
</ol>

{{if not .HasMultiplePlatforms}}
<div class="divider"></div>

<p><strong>Tip:</strong> Connect lebih banyak platform untuk melihat cross-platform analytics!</p>

<p style="text-align: center;">
    <a href="{{.BaseURL}}/dashboard/connections" class="button">Connect Platform Lain</a>
</p>
{{end}}

<p>— Tim AdsAnalytic</p>
{{end}}
`

// weeklySummaryTemplate - weekly performance digest
const weeklySummaryTemplate = `
{{define "body"}}
<h1>Laporan Mingguan Anda 📊</h1>

<p>Hai {{.Name}},</p>

<p>Ini adalah ringkasan prestasi iklan anda untuk minggu <strong>{{.WeekStart}}</strong> hingga <strong>{{.WeekEnd}}</strong>:</p>

<table width="100%" cellpadding="0" cellspacing="0" style="margin: 24px 0;">
    <tr>
        <td width="50%" style="padding: 8px;">
            <div class="metric-card">
                <div class="metric-value">RM {{.TotalSpend}}</div>
                <div class="metric-label">Jumlah Spend</div>
                <div class="metric-change {{if .SpendUp}}negative{{else}}positive{{end}}">
                    {{if .SpendUp}}↑{{else}}↓{{end}} {{.SpendChange}}% vs minggu lepas
                </div>
            </div>
        </td>
        <td width="50%" style="padding: 8px;">
            <div class="metric-card">
                <div class="metric-value">RM {{.TotalRevenue}}</div>
                <div class="metric-label">Revenue</div>
                <div class="metric-change {{if .RevenueUp}}positive{{else}}negative{{end}}">
                    {{if .RevenueUp}}↑{{else}}↓{{end}} {{.RevenueChange}}% vs minggu lepas
                </div>
            </div>
        </td>
    </tr>
    <tr>
        <td width="50%" style="padding: 8px;">
            <div class="metric-card">
                <div class="metric-value">{{.ROAS}}x</div>
                <div class="metric-label">ROAS</div>
                <div class="metric-change {{if .ROASUp}}positive{{else}}negative{{end}}">
                    {{if .ROASUp}}↑{{else}}↓{{end}} {{.ROASChange}} vs minggu lepas
                </div>
            </div>
        </td>
        <td width="50%" style="padding: 8px;">
            <div class="metric-card">
                <div class="metric-value">{{.Conversions}}</div>
                <div class="metric-label">Conversions</div>
                <div class="metric-change {{if .ConversionsUp}}positive{{else}}negative{{end}}">
                    {{if .ConversionsUp}}↑{{else}}↓{{end}} {{.ConversionsChange}}% vs minggu lepas
                </div>
            </div>
        </td>
    </tr>
</table>

<div class="divider"></div>

<h2>Prestasi Platform</h2>

{{range .Platforms}}
<div style="background-color: #f9fafb; border-radius: 8px; padding: 16px; margin: 12px 0;">
    <strong>{{.Name}}</strong>
    <div style="display: flex; justify-content: space-between; margin-top: 8px; font-size: 14px;">
        <span>Spend: RM {{.Spend}}</span>
        <span>ROAS: {{.ROAS}}x</span>
        <span>Conv: {{.Conversions}}</span>
    </div>
</div>
{{end}}

{{if .TopCampaign}}
<div class="divider"></div>

<h2>🏆 Kempen Terbaik Minggu Ini</h2>
<div class="success-box">
    <strong>{{.TopCampaign.Name}}</strong><br>
    ROAS: {{.TopCampaign.ROAS}}x | Revenue: RM {{.TopCampaign.Revenue}}
</div>
{{end}}

{{if .Insights}}
<div class="divider"></div>

<h2>💡 Insights</h2>
{{range .Insights}}
<p>• {{.}}</p>
{{end}}
{{end}}

<p style="text-align: center;">
    <a href="{{.BaseURL}}/dashboard" class="button">Lihat Dashboard Penuh</a>
</p>

<p>— Tim AdsAnalytic</p>
{{end}}
`

// tokenExpiredTemplate - token expired, need to reconnect
const tokenExpiredTemplate = `
{{define "body"}}
<h1>⚠️ Platform perlu disambung semula</h1>

<p>Hai {{.Name}},</p>

<p>Token akses untuk <strong>{{.PlatformName}}</strong> anda telah tamat tempoh. Ini bermakna kami tidak dapat sync data terkini dari platform ini.</p>

<div class="warning-box">
    <strong>Data tidak lagi di-sync!</strong><br>
    Sila sambung semula platform anda untuk meneruskan tracking.
</div>

<p>Ini biasanya berlaku kerana:</p>
<ul>
    <li>Token akses telah tamat tempoh (selepas 60 hari)</li>
    <li>Anda menukar kata laluan platform</li>
    <li>Platform membatalkan akses aplikasi</li>
</ul>

<p style="text-align: center;">
    <a href="{{.BaseURL}}/dashboard/connections" class="button">Sambung Semula {{.PlatformName}}</a>
</p>

<p>Proses ini hanya mengambil masa beberapa saat dan data anda akan mula sync semula.</p>

<p>— Tim AdsAnalytic</p>
{{end}}
`

// subscriptionConfirmTemplate - subscription confirmation
const subscriptionConfirmTemplate = `
{{define "body"}}
<h1>Terima kasih atas langganan anda! 🎉</h1>

<p>Hai {{.Name}},</p>

<p>Langganan <strong>{{.PlanName}}</strong> anda telah berjaya diaktifkan!</p>

<div class="success-box">
    <strong>Butiran Langganan</strong><br>
    Pelan: {{.PlanName}}<br>
    Harga: RM {{.Amount}}/bulan<br>
    Tarikh pembaharuan: {{.NextBillingDate}}
</div>

<h2>Apa yang anda dapat:</h2>
<ul>
{{range .Features}}
    <li>✅ {{.}}</li>
{{end}}
</ul>

<p style="text-align: center;">
    <a href="{{.BaseURL}}/dashboard" class="button">Mula Guna Sekarang</a>
</p>

<p><strong>Resit pembayaran:</strong> Resit telah dihantar ke emel anda dalam emel berasingan.</p>

<p>Terima kasih kerana memilih AdsAnalytic!</p>
<p>— Tim AdsAnalytic</p>
{{end}}
`

// paymentFailedTemplate - payment failed warning
const paymentFailedTemplate = `
{{define "body"}}
<h1>⚠️ Pembayaran gagal</h1>

<p>Hai {{.Name}},</p>

<p>Kami tidak dapat memproses pembayaran untuk langganan <strong>{{.PlanName}}</strong> anda.</p>

<div class="error-box">
    <strong>Tindakan diperlukan!</strong><br>
    Sila kemaskini maklumat pembayaran anda untuk mengelakkan gangguan perkhidmatan.
</div>

<p><strong>Butiran:</strong></p>
<ul>
    <li>Jumlah: RM {{.Amount}}</li>
    <li>Sebab: {{.FailureReason}}</li>
    <li>Percubaan: {{.AttemptCount}} / 3</li>
</ul>

{{if .GracePeriodEnd}}
<p><strong>Penting:</strong> Anda mempunyai sehingga <strong>{{.GracePeriodEnd}}</strong> untuk mengemaskini pembayaran sebelum akaun anda diturunkan ke pelan Percuma.</p>
{{end}}

<p style="text-align: center;">
    <a href="{{.BaseURL}}/dashboard/settings/billing" class="button">Kemaskini Pembayaran</a>
</p>

<p>Jika anda mempunyai sebarang soalan, sila hubungi kami di <a href="mailto:billing@adsanalytic.com">billing@adsanalytic.com</a>.</p>

<p>— Tim AdsAnalytic</p>
{{end}}
`

// connectReminderTemplate - 24h no platform connected reminder
const connectReminderTemplate = `
{{define "body"}}
<h1>Jangan lupa connect platform anda! 📱</h1>

<p>Hai {{.Name}},</p>

<p>Kami perasan anda belum menyambungkan sebarang platform iklan ke AdsAnalytic. Untuk mula melihat analitik, anda perlu connect sekurang-kurangnya satu platform.</p>

<p>Ia hanya mengambil masa <strong>kurang dari 2 minit</strong>!</p>

<h2>Platform yang disokong:</h2>
<table width="100%" cellpadding="0" cellspacing="0">
    <tr>
        <td width="33%" style="padding: 8px; text-align: center;">
            <div style="background: #1877f2; color: white; padding: 16px; border-radius: 8px;">
                <strong>Meta Ads</strong><br>
                <small>Facebook & Instagram</small>
            </div>
        </td>
        <td width="33%" style="padding: 8px; text-align: center;">
            <div style="background: #000000; color: white; padding: 16px; border-radius: 8px;">
                <strong>TikTok Ads</strong><br>
                <small>TikTok For Business</small>
            </div>
        </td>
        <td width="33%" style="padding: 8px; text-align: center;">
            <div style="background: #ee4d2d; color: white; padding: 16px; border-radius: 8px;">
                <strong>Shopee Ads</strong><br>
                <small>Shopee My Ads</small>
            </div>
        </td>
    </tr>
</table>

<p style="text-align: center; margin-top: 24px;">
    <a href="{{.BaseURL}}/dashboard/connections" class="button">Connect Platform Sekarang</a>
</p>

<p>Perlu bantuan? Layari <a href="{{.BaseURL}}/docs/getting-started">panduan permulaan</a> kami.</p>

<p>— Tim AdsAnalytic</p>
{{end}}
`

// inactiveReminderTemplate - 3 days inactive reminder
const inactiveReminderTemplate = `
{{define "body"}}
<h1>Data iklan anda menunggu! 📊</h1>

<p>Hai {{.Name}},</p>

<p>Kami perasan anda tidak log masuk ke AdsAnalytic sejak <strong>{{.LastLoginDays}} hari</strong> yang lalu.</p>

<p>Banyak yang berlaku dengan iklan anda:</p>

{{if .HasData}}
<div class="metric-card">
    <div class="metric-label">Sejak lawatan terakhir anda</div>
    <div style="margin-top: 8px;">
        <span style="margin-right: 16px;">💰 Spend: RM {{.RecentSpend}}</span>
        <span>📈 Conversions: {{.RecentConversions}}</span>
    </div>
</div>
{{end}}

<p style="text-align: center;">
    <a href="{{.BaseURL}}/dashboard" class="button">Lihat Dashboard</a>
</p>

<p>Tip: Set up <a href="{{.BaseURL}}/dashboard/settings/notifications">weekly email digest</a> supaya anda sentiasa updated tanpa perlu log masuk!</p>

<p>— Tim AdsAnalytic</p>
{{end}}
`

// GetSubject returns the default subject for an email type
func GetSubject(emailType EmailType, data map[string]interface{}) string {
	switch emailType {
	case EmailTypeWelcome:
		return "Selamat datang ke AdsAnalytic! 🎉"
	case EmailTypeVerification:
		return "Sahkan alamat emel anda - AdsAnalytic"
	case EmailTypePasswordReset:
		return "Reset kata laluan anda - AdsAnalytic"
	case EmailTypePlatformConnected:
		platform := data["PlatformName"]
		return fmt.Sprintf("🎊 %v berjaya disambungkan!", platform)
	case EmailTypeWeeklySummary:
		return "📊 Laporan Mingguan Anda - AdsAnalytic"
	case EmailTypeTokenExpired:
		platform := data["PlatformName"]
		return fmt.Sprintf("⚠️ %v perlu disambung semula", platform)
	case EmailTypeSubscriptionConfirm:
		plan := data["PlanName"]
		return fmt.Sprintf("Terima kasih! Langganan %v aktif 🎉", plan)
	case EmailTypePaymentFailed:
		return "⚠️ Pembayaran gagal - Tindakan diperlukan"
	case EmailTypeConnectReminder:
		return "Jangan lupa connect platform iklan anda! 📱"
	case EmailTypeInactiveReminder:
		return "Data iklan anda menunggu! 📊"
	default:
		return "Notifikasi dari AdsAnalytic"
	}
}
