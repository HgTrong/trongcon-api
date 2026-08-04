package service

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"trongcon-api/internal/config"
)

type VNPayCreateResult struct {
	TxnRef     string
	PaymentURL string
}

type VNPayVerifyResult struct {
	Valid           bool
	Success         bool
	TxnRef          string
	TransactionNo   string
	ResponseCode    string
	AmountVND       int64
	Message         string
}

type VNPayService interface {
	Enabled() bool
	CreatePaymentURL(txnRef string, amountVND int64, orderInfo string, ipAddr string, returnURL string) (*VNPayCreateResult, error)
	VerifyReturn(params map[string]string) (*VNPayVerifyResult, error)
	AmountVND(price float64, currency string) int64
}

type vnPayService struct {
	cfg config.VNPayConfig
}

func NewVNPayService(cfg config.VNPayConfig) VNPayService {
	return &vnPayService{cfg: cfg}
}

func (s *vnPayService) Enabled() bool {
	return strings.TrimSpace(s.cfg.TmnCode) != "" && strings.TrimSpace(s.cfg.SecretKey) != ""
}

// AmountVND returns the catalog price as VND integer (before *100 for vnp_Amount).
// TrongCon prices are VND-only; currency is ignored.
func (s *vnPayService) AmountVND(price float64, currency string) int64 {
	_ = currency
	return int64(math.Round(price))
}

func (s *vnPayService) CreatePaymentURL(txnRef string, amountVND int64, orderInfo string, ipAddr string, returnURL string) (*VNPayCreateResult, error) {
	if !s.Enabled() {
		return nil, fmt.Errorf("vnpay is not configured")
	}
	txnRef = strings.TrimSpace(txnRef)
	if txnRef == "" {
		return nil, fmt.Errorf("txn_ref required")
	}
	if amountVND < 1000 {
		return nil, fmt.Errorf("amount too small for VNPay (min ~1000 VND)")
	}
	if ipAddr == "" {
		ipAddr = "127.0.0.1"
	}
	orderInfo = strings.TrimSpace(orderInfo)
	if orderInfo == "" {
		orderInfo = "TrongCon Premium"
	}
	// VNPay rejects some special chars in OrderInfo
	orderInfo = sanitizeVNPayText(orderInfo)

	returnURL = strings.TrimSpace(returnURL)
	if returnURL == "" {
		returnURL = strings.TrimSpace(s.cfg.ReturnURL)
	}

	loc := time.FixedZone("Asia/Ho_Chi_Minh", 7*3600)
	now := time.Now().In(loc)
	createDate := now.Format("20060102150405")
	expireDate := now.Add(15 * time.Minute).Format("20060102150405")

	params := map[string]string{
		"vnp_Version":    "2.1.0",
		"vnp_Command":    "pay",
		"vnp_TmnCode":    strings.TrimSpace(s.cfg.TmnCode),
		"vnp_Amount":     strconv.FormatInt(amountVND*100, 10),
		"vnp_CurrCode":   "VND",
		"vnp_TxnRef":     txnRef,
		"vnp_OrderInfo":  orderInfo,
		"vnp_OrderType":  "other",
		"vnp_Locale":     "vn",
		"vnp_ReturnUrl":  returnURL,
		"vnp_IpAddr":     ipAddr,
		"vnp_CreateDate": createDate,
		"vnp_ExpireDate": expireDate,
	}

	// Hash must use the same URL-encoded query string VNPay rebuilds (PHP demo).
	signData := buildVNPaySignData(params)
	secureHash := hmacSHA512(strings.TrimSpace(s.cfg.SecretKey), signData)

	payURL := strings.TrimRight(s.cfg.PaymentURL, "?") + "?" + signData + "&vnp_SecureHash=" + secureHash
	return &VNPayCreateResult{TxnRef: txnRef, PaymentURL: payURL}, nil
}

func (s *vnPayService) VerifyReturn(params map[string]string) (*VNPayVerifyResult, error) {
	if !s.Enabled() {
		return nil, fmt.Errorf("vnpay is not configured")
	}
	secureHash := strings.TrimSpace(params["vnp_SecureHash"])
	if secureHash == "" {
		return nil, fmt.Errorf("missing vnp_SecureHash")
	}

	check := make(map[string]string, len(params))
	for k, v := range params {
		if k == "vnp_SecureHash" || k == "vnp_SecureHashType" {
			continue
		}
		if strings.HasPrefix(k, "vnp_") {
			check[k] = v
		}
	}
	signData := buildVNPaySignData(check)
	expected := hmacSHA512(s.cfg.SecretKey, signData)
	valid := strings.EqualFold(expected, secureHash)

	code := params["vnp_ResponseCode"]
	txnRef := params["vnp_TxnRef"]
	transNo := params["vnp_TransactionNo"]
	amountRaw, _ := strconv.ParseInt(params["vnp_Amount"], 10, 64)

	msg := "payment failed"
	success := valid && code == "00"
	if !valid {
		msg = "invalid signature"
	} else if code == "00" {
		msg = "success"
	} else if code == "24" {
		msg = "customer cancelled"
	} else {
		msg = "response_code=" + code
	}

	return &VNPayVerifyResult{
		Valid:         valid,
		Success:       success,
		TxnRef:        txnRef,
		TransactionNo: transNo,
		ResponseCode:  code,
		AmountVND:     amountRaw / 100,
		Message:       msg,
	}, nil
}

func buildVNPaySignData(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if strings.TrimSpace(v) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		// Match VNPay PHP: urlencode(key)=urlencode(value)
		parts = append(parts, vnpayEscape(k)+"="+vnpayEscape(params[k]))
	}
	return strings.Join(parts, "&")
}

// vnpayEscape matches PHP urlencode / Java URLEncoder (space → +).
func vnpayEscape(s string) string {
	return url.QueryEscape(s)
}

func hmacSHA512(key, data string) string {
	h := hmac.New(sha512.New, []byte(key))
	_, _ = h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func sanitizeVNPayText(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	replacer := strings.NewReplacer("|", " ", "&", " ", "=", " ", "?", " ")
	s = replacer.Replace(s)
	if len(s) > 255 {
		s = s[:255]
	}
	return s
}

// NewVNPayTxnRef builds a unique txn ref (max ~100 chars for VNPay).
func NewVNPayTxnRef(subID uint) string {
	return NewVNPayTxnRefWithPrefix("TC", subID)
}

// NewVNPayTxnRefWithPrefix builds a unique, prefixed txn ref, e.g. GM{id}T{ms}.
func NewVNPayTxnRefWithPrefix(prefix string, id uint) string {
	return fmt.Sprintf("%s%dT%d", prefix, id, time.Now().UnixNano()/1e6)
}
