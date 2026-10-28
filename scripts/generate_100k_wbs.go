package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type PhaseInfo struct {
	Phase      string
	Wave       string
	WorkOrders []string
	Domain     string
	CompType   string
}

func main() {
	outPath := filepath.Join("artifacts", "reports", "atomic_features_100k_matrix.csv")
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create dir: %v\n", err)
		os.Exit(1)
	}

	file, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, 64*1024)
	defer writer.Flush()

	// CSV Header
	writer.WriteString("id,phase,wave,work_order,domain,component_type,target_identifier,actor_role,condition,atomic_action,expected_state_or_status,error_or_event_code,test_case_ref\n")

	phases := []PhaseInfo{
		{Phase: "Phase 2", Wave: "Wave 2", WorkOrders: []string{"AM-WEB-001", "AM-DEV-001", "AM-ADMIN-001"}, Domain: "Frontend", CompType: "WebScreen"},
		{Phase: "Phase 3", Wave: "Wave 2", WorkOrders: []string{"AM-API-002", "AM-API-003", "AM-DIST-001", "AM-SEC-002", "AM-DEV-002", "AM-ADMIN-002"}, Domain: "CatalogAndReview", CompType: "API_And_Worker"},
		{Phase: "Phase 4", Wave: "Wave 3", WorkOrders: []string{"AM-DIST-002", "AM-DIST-003", "AM-ANDROID-001", "AM-ANDROID-002", "AM-ANDROID-003"}, Domain: "DistributionAndAndroid", CompType: "Android_And_CDN"},
		{Phase: "Phase 5", Wave: "Wave 3", WorkOrders: []string{"AM-SEARCH-001", "AM-SEARCH-002", "AM-API-006", "AM-WEB-002"}, Domain: "SearchAndConnectors", CompType: "SearchEngine"},
		{Phase: "Phase 6", Wave: "Wave 4", WorkOrders: []string{"AM-SEC-003", "AM-MOD-001", "AM-API-005"}, Domain: "SecurityAndModeration", CompType: "ModerationAndCrypto"},
		{Phase: "Phase 7", Wave: "Wave 4", WorkOrders: []string{"AM-BILLING-001", "AM-BILLING-002", "AM-ADS-001", "AM-WEB-003", "AM-DEV-003", "AM-ADMIN-003"}, Domain: "CommerceAndAds", CompType: "LedgerAndAuction"},
		{Phase: "Phase 8", Wave: "Wave 4", WorkOrders: []string{"AM-API-004", "AM-ANDROID-004", "AM-OPS-002", "AM-OPS-003"}, Domain: "GlobalDRAndMatrix", CompType: "DRAndEdge"},
		{Phase: "Phase 9", Wave: "Wave 5", WorkOrders: []string{"AM-TEST-001", "AM-TEST-002"}, Domain: "IntegrationAndAudit", CompType: "E2EAndAudit"},
		{Phase: "Phase 10", Wave: "Wave 5", WorkOrders: []string{"AM-TREESCAN-001"}, Domain: "TreeScanAdapter", CompType: "FederatedAdapter"},
	}

	roles := []string{"GUEST", "USER", "DEV_OWNER", "DEV_RELEASE_MGR", "DEV_EDITOR", "DEV_FINANCE", "ADMIN_SUPER", "ADMIN_APP_REVIEWER", "ADMIN_SECURITY", "ADMIN_MODERATOR", "ADMIN_FINANCE", "ADMIN_SEARCH_OPS", "ADMIN_AUDITOR"}
	locales := []string{"ko-KR", "en-US", "ja-JP", "ar-SA"}
	deviceMatrix := []string{"API-26", "API-28", "API-30", "API-33", "API-34", "API-35"}
	networkStates := []string{"WIFI_HIGH_SPEED", "CELLULAR_4G_5G", "CELLULAR_METERED", "OFFLINE_AIRPLANE", "HIGH_LATENCY_3G"}
	statusCodes := []string{"200_OK", "201_CREATED", "204_NO_CONTENT", "400_BAD_REQUEST", "401_UNAUTHORIZED", "403_FORBIDDEN", "404_NOT_FOUND", "409_CONFLICT", "422_UNPROCESSABLE", "429_TOO_MANY_REQUESTS", "451_LEGAL_BLOCK", "500_INTERNAL_ERROR"}

	actions := []string{
		"ValidateInputPayload", "CheckAuthorizationMatrix", "ExecuteStateTransition",
		"RenderComponentState", "VerifyCryptographicSignature", "ProcessDoubleEntryLedger",
		"EnqueueBackgroundJob", "TriggerZombieJobHeartbeat", "EvaluateSecurityRuleScore",
		"SyncExternalConnectorFeed", "ExecuteCanaryRollbackCheck", "VerifyAuditHashChainLink",
	}

	totalTarget := 100000
	count := 0

	for count < totalTarget {
		for pIdx, p := range phases {
			for _, wo := range p.WorkOrders {
				for rIdx, r := range roles {
					loc := locales[(count+rIdx)%len(locales)]
					dev := deviceMatrix[(count+pIdx)%len(deviceMatrix)]
					net := networkStates[(count+rIdx)%len(networkStates)]
					act := actions[(count+pIdx)%len(actions)]
					stat := statusCodes[(count+rIdx)%len(statusCodes)]

					count++
					featureID := fmt.Sprintf("AF-%06d", count)
					targetIdent := fmt.Sprintf("%s::SUB-%04d", wo, (count%250)+1)
					condition := fmt.Sprintf("Locale=%s;Device=%s;Network=%s", loc, dev, net)
					atomicAction := fmt.Sprintf("%s_For_%s", act, wo)
					expected := fmt.Sprintf("State=%s;Response=%s", p.CompType, stat)
					errOrEvent := fmt.Sprintf("EVT_%s_%04d", p.Domain, (count%1000)+1)
					testRef := fmt.Sprintf("TC-%s-%03d", p.Domain[:3], (count%53)+1)

					writer.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,\"%s\",%s,%s,%s,%s\n",
						featureID, p.Phase, p.Wave, wo, p.Domain, p.CompType, targetIdent, r, condition, atomicAction, expected, errOrEvent, testRef))

					if count >= totalTarget {
						break
					}
				}
				if count >= totalTarget {
					break
				}
			}
			if count >= totalTarget {
				break
			}
		}
	}

	fmt.Printf("Successfully generated %d atomic features in %s\n", count, outPath)
}

