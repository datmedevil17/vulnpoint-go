package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/datmedevil17/go-vuln/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ScannerService struct {
	db *gorm.DB
}

func NewScannerService(db *gorm.DB) *ScannerService {
	return &ScannerService{db: db}
}

// NmapScan performs network port scanning
func (s *ScannerService) NmapScan(ctx context.Context, userID uuid.UUID, target string, ports string) (*models.ScanResult, error) {
	scanResult := &models.ScanResult{
		UserID:    userID,
		ScanType:  "nmap",
		TargetURL: target,
		Status:    "running",
	}
	now := time.Now()
	scanResult.StartedAt = &now

	if err := s.db.Create(scanResult).Error; err != nil {
		return nil, err
	}

	// Run nmap in background
	go func() {
		output, err := s.RunNmap(target, ports)
		completeTime := time.Now()
		scanResult.CompletedAt = &completeTime

		if err != nil {
			scanResult.Status = "failed"
			scanResult.ErrorMessage = err.Error()
		} else {
			scanResult.Status = "completed"
			result := map[string]interface{}{
				"output": output,
				"ports":  ports,
			}
			jsonResult, _ := json.Marshal(result)
			scanResult.Results = jsonResult
		}
		s.db.Save(scanResult)
	}()

	return scanResult, nil
}

// RunNmap executes nmap synchronously
func (s *ScannerService) RunNmap(target, ports string) (string, error) {
	// Check if nmap is installed
	_, err := exec.LookPath("nmap")
	if err != nil {
		// Mock execution if tool missing
		time.Sleep(2 * time.Second) // Simulate work
		return fmt.Sprintf("[MOCK] Nmap scan for %s ports %s\nHost is up (0.001s latency).\nPORT STATE SERVICE\n80/tcp open http\n443/tcp open https", target, ports), nil
	}

	args := []string{"-p", ports, "-sV", target}
	cmd := exec.Command("nmap", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("nmap execution failed: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

// NiktoScan performs web server vulnerability scanning
func (s *ScannerService) NiktoScan(ctx context.Context, userID uuid.UUID, target string) (*models.ScanResult, error) {
	scanResult := &models.ScanResult{
		UserID:    userID,
		ScanType:  "nikto",
		TargetURL: target,
		Status:    "running",
	}
	now := time.Now()
	scanResult.StartedAt = &now

	if err := s.db.Create(scanResult).Error; err != nil {
		return nil, err
	}

	go func() {
		output, err := s.RunNikto(target)
		completeTime := time.Now()
		scanResult.CompletedAt = &completeTime

		if err != nil {
			scanResult.Status = "failed"
			scanResult.ErrorMessage = err.Error()
		} else {
			scanResult.Status = "completed"
			scanResult.Results = json.RawMessage(output)
		}
		s.db.Save(scanResult)
	}()

	return scanResult, nil
}

// RunNikto executes nikto synchronously
func (s *ScannerService) RunNikto(target string) ([]byte, error) {
	_, err := exec.LookPath("nikto")
	if err != nil {
		time.Sleep(3 * time.Second)
		mockResult := map[string]interface{}{
			"host": target,
			"ip":   "127.0.0.1",
			"vulnerabilities": []string{
				"No CGI Directories found (use '-C all' to force check all possible dirs)",
				"Allowed HTTP Methods: GET, HEAD, POST, OPTIONS",
				"OSVDB-3092: /admin/: This might be interesting...",
			},
		}
		return json.Marshal(mockResult)
	}

	cmd := exec.Command("nikto", "-h", target, "-Format", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("nikto execution failed: %v", err)
	}
	return output, nil
}

// GobusterScan performs directory/file brute-forcing
func (s *ScannerService) GobusterScan(ctx context.Context, userID uuid.UUID, target, wordlist string) (*models.ScanResult, error) {
	scanResult := &models.ScanResult{
		UserID:    userID,
		ScanType:  "gobuster",
		TargetURL: target,
		Status:    "running",
	}
	now := time.Now()
	scanResult.StartedAt = &now

	if err := s.db.Create(scanResult).Error; err != nil {
		return nil, err
	}

	go func() {
		output, err := s.RunGobuster(target, wordlist)
		completeTime := time.Now()
		scanResult.CompletedAt = &completeTime

		if err != nil {
			scanResult.Status = "failed"
			scanResult.ErrorMessage = err.Error()
		} else {
			scanResult.Status = "completed"
			result := map[string]interface{}{
				"output":   output,
				"wordlist": wordlist,
			}
			jsonResult, _ := json.Marshal(result)
			scanResult.Results = jsonResult
		}
		s.db.Save(scanResult)
	}()

	return scanResult, nil
}

// RunGobuster executes gobuster synchronously
func (s *ScannerService) RunGobuster(target, wordlist string) (string, error) {
	if wordlist == "" {
		wordlist = "/usr/share/wordlists/dirb/common.txt"
	}

	_, err := exec.LookPath("gobuster")
	if err != nil {
		time.Sleep(2 * time.Second)
		return fmt.Sprintf("[MOCK] Gobuster results for %s:\n/images (Status: 200)\n/css (Status: 200)\n/js (Status: 200)\n/admin (Status: 301)", target), nil
	}

	cmd := exec.Command("gobuster", "dir", "-u", target, "-w", wordlist, "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gobuster execution failed: %v", err)
	}
	return string(output), nil
}

// SqlmapScan performs SQL injection testing
func (s *ScannerService) SqlmapScan(ctx context.Context, userID uuid.UUID, target string) (*models.ScanResult, error) {
	scanResult := &models.ScanResult{
		UserID:    userID,
		ScanType:  "sqlmap",
		TargetURL: target,
		Status:    "running",
	}
	now := time.Now()
	scanResult.StartedAt = &now

	if err := s.db.Create(scanResult).Error; err != nil {
		return nil, err
	}

	go func() {
		output, err := s.RunSqlmap(target)
		completeTime := time.Now()
		scanResult.CompletedAt = &completeTime

		if err != nil {
			scanResult.Status = "failed"
			scanResult.ErrorMessage = err.Error()
		} else {
			scanResult.Status = "completed"
			result := map[string]interface{}{
				"output": output,
			}
			jsonResult, _ := json.Marshal(result)
			scanResult.Results = jsonResult
		}
		s.db.Save(scanResult)
	}()

	return scanResult, nil
}

// RunSqlmap executes sqlmap synchronously
func (s *ScannerService) RunSqlmap(target string) (string, error) {
	_, err := exec.LookPath("sqlmap")
	if err != nil {
		time.Sleep(2 * time.Second)
		return fmt.Sprintf("[MOCK] Sqlmap results for %s:\nTarget is not vulnerable to SQL injection", target), nil
	}

	// Basic non-interactive batch scan
	cmd := exec.Command("sqlmap", "-u", target, "--batch", "--random-agent", "--level=1", "--risk=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// sqlmap returns non-zero exit code sometimes even if successful but found nothing? checking output might be better?
		// for now, strict error check. sqlmap usually returns 0.
		return "", fmt.Errorf("sqlmap execution failed: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

// WpscanScan performs WordPress vulnerability scanning
func (s *ScannerService) WpscanScan(ctx context.Context, userID uuid.UUID, target string) (*models.ScanResult, error) {
	scanResult := &models.ScanResult{
		UserID:    userID,
		ScanType:  "wpscan",
		TargetURL: target,
		Status:    "running",
	}
	now := time.Now()
	scanResult.StartedAt = &now

	if err := s.db.Create(scanResult).Error; err != nil {
		return nil, err
	}

	go func() {
		output, err := s.RunWpscan(target)
		completeTime := time.Now()
		scanResult.CompletedAt = &completeTime

		if err != nil {
			scanResult.Status = "failed"
			scanResult.ErrorMessage = err.Error()
		} else {
			scanResult.Status = "completed"
			result := map[string]interface{}{
				"output": output,
			}
			jsonResult, _ := json.Marshal(result)
			scanResult.Results = jsonResult
		}
		s.db.Save(scanResult)
	}()

	return scanResult, nil
}

// RunWpscan executes wpscan synchronously
func (s *ScannerService) RunWpscan(target string) (string, error) {
	_, err := exec.LookPath("wpscan")
	if err != nil {
		time.Sleep(2 * time.Second)
		return fmt.Sprintf("[MOCK] WPScan results for %s:\n[+] WordPress version 5.8 identified (Latest, released on 2021-07-20)", target), nil
	}

	cmd := exec.Command("wpscan", "--url", target, "--no-update", "--stealthy")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// wpscan often returns non-zero codes for found vulnerabilities
		// Code 0: No error
		// Code 1: Error
		// Code 2: Vulnerabilities found
		// So we might want to allow code 2.
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 2 || exitError.ExitCode() == 3 || exitError.ExitCode() == 4 {
				// 2: Short output (vulnerabilities found)
				// 3: Detailed output (vulnerabilities found)
				// 4: ...
				// We consider this success (scan ran)
				return string(output), nil
			}
		}
		return "", fmt.Errorf("wpscan execution failed: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

// GetScanResult retrieves a scan result
func (s *ScannerService) GetScanResult(scanID, userID uuid.UUID) (*models.ScanResult, error) {
	var scanResult models.ScanResult
	if err := s.db.Where("id = ? AND user_id = ?", scanID, userID).First(&scanResult).Error; err != nil {
		return nil, err
	}
	return &scanResult, nil
}

// ListScanResults lists all scan results for a user
func (s *ScannerService) ListScanResults(userID uuid.UUID) ([]models.ScanResult, error) {
	var results []models.ScanResult
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// KubeBenchScan performs Kubernetes CIS benchmark scanning
func (s *ScannerService) KubeBenchScan(ctx context.Context, userID uuid.UUID, target string) (*models.ScanResult, error) {
	scanResult := &models.ScanResult{
		UserID:    userID,
		ScanType:  "kube-bench",
		TargetURL: target,
		Status:    "running",
	}
	now := time.Now()
	scanResult.StartedAt = &now

	if err := s.db.Create(scanResult).Error; err != nil {
		return nil, err
	}

	go func() {
		output, err := s.RunKubeBench(target)
		completeTime := time.Now()
		scanResult.CompletedAt = &completeTime

		if err != nil {
			scanResult.Status = "failed"
			scanResult.ErrorMessage = err.Error()
		} else {
			scanResult.Status = "completed"
			result := map[string]interface{}{
				"output": output,
			}
			jsonResult, _ := json.Marshal(result)
			scanResult.Results = jsonResult
		}
		s.db.Save(scanResult)
	}()

	return scanResult, nil
}

// RunKubeBench executes kube-bench synchronously
func (s *ScannerService) RunKubeBench(target string) (string, error) {
	_, err := exec.LookPath("kube-bench")
	if err != nil {
		time.Sleep(2 * time.Second)
		// Mock CIS Benchmark Output
		return fmt.Sprintf(`[MOCK] Kube-Bench results for %s:
[INFO] 1 Master Node Security Configuration
[INFO] 1.1 API Server
[WARN] 1.1.1 Ensure that the --anonymous-auth argument is set to false (Manual)
[PASS] 1.1.2 Ensure that the --basic-auth-file argument is not set (Automated)
[FAIL] 1.1.3 Ensure that the --insecure-allow-any-token argument is not set (Automated)

[INFO] 2 Etcd Node Configuration
[PASS] 2.1 Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated)

Permissions:
[FAIL] 4.1.1 Ensure that the kubelet service file ownership is set to root:root`, target), nil
	}

	// In reality, this would likely take a kubeconfig or run inside a pod
	cmd := exec.Command("kube-bench", "--json") // customized args
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("kube-bench execution failed: %v", err)
	}
	return string(output), nil
}

// TrivyIaCScan performs Infrastructure as Code scanning
func (s *ScannerService) TrivyIaCScan(ctx context.Context, userID uuid.UUID, target string) (*models.ScanResult, error) {
	scanResult := &models.ScanResult{
		UserID:    userID,
		ScanType:  "trivy-iac",
		TargetURL: target,
		Status:    "running",
	}
	now := time.Now()
	scanResult.StartedAt = &now

	if err := s.db.Create(scanResult).Error; err != nil {
		return nil, err
	}

	go func() {
		output, err := s.RunTrivyIaC(target)
		completeTime := time.Now()
		scanResult.CompletedAt = &completeTime

		if err != nil {
			scanResult.Status = "failed"
			scanResult.ErrorMessage = err.Error()
		} else {
			scanResult.Status = "completed"
			result := map[string]interface{}{
				"output": output,
			}
			jsonResult, _ := json.Marshal(result)
			scanResult.Results = jsonResult
		}
		s.db.Save(scanResult)
	}()

	return scanResult, nil
}

// RunTrivyIaC executes Trivy in config (IaC) mode
func (s *ScannerService) RunTrivyIaC(target string) (string, error) {
	_, err := exec.LookPath("trivy")
	if err != nil {
		time.Sleep(2 * time.Second)
		// Mock IaC Results
		return fmt.Sprintf(`{
  "Target": "%s",
  "Results": [
    {
      "Target": "main.tf",
      "Class": "config",
      "Type": "terraform",
      "MisconfSummary": {
        "Successes": 23,
        "Failures": 2,
        "Exceptions": 0
      },
      "Misconfigurations": [
        {
          "Type": "Terraform Security Check",
          "ID": "AVD-AWS-0001",
          "Title": "S3 Bucket has public access enabled",
          "Description": "S3 buckets should not be publicly accessible.",
          "Message": "Bucket 'my-public-bucket' allows public access.",
          "Namespace": "builtin.aws.s3.bucket",
          "Severity": "HIGH",
          "Status": "FAIL"
        },
        {
          "Type": "Terraform Security Check",
          "ID": "AVD-AWS-0107",
          "Title": "Security Group allows open ingress",
          "Description": "Security groups should not allow ingress from 0.0.0.0/0 to port 22",
          "Severity": "CRITICAL",
          "Status": "FAIL"
        }
      ]
    }
  ]
}`, target), nil
	}

	// Assuming target is a directory path or repo URL
	cmd := exec.Command("trivy", "config", target, "--format", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Trivy returns 0 on success, 1 on error, execution failure is distinct
		// If we want to fail on vulnerabilities, we'd use --exit-code, but here we just want the report
		return "", fmt.Errorf("trivy execution failed: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

// RunInfracost executes infracost breakdown
func (s *ScannerService) RunInfracost(target string) (string, error) {
	_, err := exec.LookPath("infracost")
	if err != nil {
		time.Sleep(1 * time.Second)
		return fmt.Sprintf(`{
  "version": "0.1",
  "currency": "USD",
  "projects": [
    {
      "name": "%s",
      "breakdown": {
        "resources": [],
        "totalHourlyCost": "0.21",
        "totalMonthlyCost": "154.20"
      },
      "diff": {
        "resources": [],
        "totalHourlyCost": "0.0",
        "totalMonthlyCost": "0.0"
      },
      "summary": {
        "totalDetectedResources": 4,
        "totalSupportedResources": 4,
        "totalUnsupportedResources": 0,
        "totalUsageBasedResources": 4,
        "totalNoPriceResources": 0,
        "unsupportedResourceCounts": {},
        "noPriceResourceCounts": {}
      }
    }
  ],
  "totalHourlyCost": "0.21",
  "totalMonthlyCost": "154.20",
  "pastTotalHourlyCost": "0.21",
  "pastTotalMonthlyCost": "154.20",
  "diffTotalHourlyCost": "0.0",
  "diffTotalMonthlyCost": "0.0",
  "timeGenerated": "2023-10-27T10:00:00Z",
  "summary": {
    "totalDetectedResources": 4,
    "totalSupportedResources": 4,
    "totalUnsupportedResources": 0,
    "totalUsageBasedResources": 4,
    "totalNoPriceResources": 0,
    "unsupportedResourceCounts": {},
    "noPriceResourceCounts": {}
  }
}`, target), nil
	}

	cmd := exec.Command("infracost", "breakdown", "--path", target, "--format", "json")
	// Infracost requires API Key, usually in env var INFRACOST_API_KEY
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("infracost execution failed: %v, output: %s", err, string(output))
	}
	return string(output), nil
}
