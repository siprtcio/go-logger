package logger

import (
	"fmt"
	"github.com/evalphobia/logrus_sentry"
	"github.com/go-resty/resty/v2"
	"github.com/go-xmlfmt/xmlfmt"
	"github.com/sirupsen/logrus"
	"log/syslog"
	"os"
	"strconv"
)

var logLevel = map[string]logrus.Level{
	"debug": logrus.DebugLevel,
	"info":  logrus.InfoLevel,
	"warn":  logrus.WarnLevel,
}

var facilityLevel = map[string]syslog.Priority{
	"local0": syslog.LOG_LOCAL0,
	"local1": syslog.LOG_LOCAL1,
	"local2": syslog.LOG_LOCAL2,
	"local3": syslog.LOG_LOCAL3,
}

type SipRtcLogger struct {
	LogLevel        string
	LoggingFacility string
	LoggingTag      string
}

var Logger *logrus.Logger

func (sLg *SipRtcLogger) InitLogger() {
	var err error
	Logger, err = sLg.NewLogger(sLg.LogLevel, sLg.LoggingFacility, sLg.LoggingTag, "", "")
	if err != nil || Logger == nil {
		return
	}

	l := logrus.JSONFormatter{}
	l.DisableHTMLEscape = true

	Logger.SetFormatter(&l)
	Logger.SetReportCaller(true)
	Logger.AddHook(&ErrorHook{})
}

func (sLg *SipRtcLogger) NewLogFields(fieldMap map[string]interface{}) logrus.Fields {
	return logrus.Fields(fieldMap)
}

func (sLg *SipRtcLogger) GuardCritical(msg string, err error) {
	if err != nil {
		fmt.Printf("CRITICAL: %s: %v\n", msg, err)
		os.Exit(-1)
	}
}

func (sLg *SipRtcLogger) NewLogger(level, facility, tag string, sentry string, syslogAddr string) (*logrus.Logger, error) {
	l := logrus.New()

	ll, ok := logLevel[level]
	if !ok {
		ll = logLevel["debug"]
	}
	l.Level = ll

	if sentry != "" {
		hostname, err := os.Hostname()
		sLg.GuardCritical("determining hostname failed", err)

		tags := map[string]string{
			"tag":      tag,
			"hostname": hostname,
		}

		sentryLevels := []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		}
		sentHook, err := logrus_sentry.NewWithTagsSentryHook(sentry, tags, sentryLevels)
		sLg.GuardCritical("configuring sentry failed", err)

		l.Hooks.Add(sentHook)
	}

	return l, nil
}

func (sLg *SipRtcLogger) BuildLogEntry(l *logrus.Entry, in map[string]string) *logrus.Entry {
	for k, v := range in {
		l = l.WithField(k, v)
	}
	return l
}

func (sLg *SipRtcLogger) XmlLog(logLevel, uuid, message string) {
	xmlMsg := xmlfmt.FormatXML(message, "", "  ", true)
	if logLevel == "Err" {
		Logger.WithField("requestId", uuid).Error(xmlMsg)
	} else if logLevel == "Info" {
		Logger.WithField("requestId", uuid).Info(xmlMsg)
	} else {
		Logger.WithField("requestId", uuid).Debug(xmlMsg)
	}
}

func (sLg *SipRtcLogger) CallLog(logLevel, uuid, direction, message string) {

	if unQuoteMsg, err := strconv.Unquote(message); err == nil {
		message = unQuoteMsg
	}

	if logLevel == "Err" {
		Logger.WithField("requestId", uuid).WithField("direction", direction).Error(message)
	} else if logLevel == "Info" {
		Logger.WithField("requestId", uuid).WithField("direction", direction).Info(message)
	}
	Logger.WithField("requestId", uuid).WithField("direction", direction).Debug(message)
}

func (sLg *SipRtcLogger) HttpTraceLog(logLevel, requestId string, resp *resty.Response) {
	if resp != nil {
		ti := resp.Request.TraceInfo()
		Logger.WithField("requestId", requestId).WithField("Status", resp.Status()).
			WithField("DNSLookup", ti.DNSLookup).
			WithField("ConnTime", ti.ConnTime).
			WithField("TCPConnTime", ti.TCPConnTime).
			WithField("TLSHandshake", ti.TLSHandshake).
			WithField("ServerTime", ti.ServerTime).
			WithField("ResponseTime", ti.ResponseTime).
			WithField("TotalTime", ti.TotalTime).
			WithField("IsConnReused", ti.IsConnReused).
			WithField("IsConnWasIdle", ti.IsConnWasIdle).
			WithField("ConnIdleTime", ti.ConnIdleTime).
			WithField("RequestAttempt", ti.RequestAttempt).
			Info("Http Response Received")
	}
}
