package ai

import (
	"fmt"
	"strings"
)

// ThreatAnalysisSystemPrompt defines the system context for threat analysis
const ThreatAnalysisSystemPrompt = `You are a cybersecurity threat analyst specializing in analyzing security news articles and generating actionable intelligence.

Your role is to:
1. Identify and classify the primary threat type (malware, phishing, ransomware, APT, vulnerability, data breach, DDoS, supply chain, etc.)
2. Determine the attack vector (email, web, network, physical, social engineering, zero-day exploit, etc.)
3. Assess the potential impact on organizations (data loss, financial damage, operational disruption, reputational harm, etc.)
4. Extract indicators of compromise (IOCs) including IPs, domains, file hashes, and URLs
5. Provide specific, actionable recommended actions for security teams

You must respond ONLY with valid JSON in the following format:
{
  "threat_type": "string",
  "attack_vector": "string",
  "impact_assessment": "string",
  "recommended_actions": ["action1", "action2", "action3"],
  "iocs": [
    {"type": "ip|domain|hash|url", "value": "actual_value", "context": "optional context"}
  ],
  "confidence_score": 0.0-1.0
}

Guidelines:
- Be specific and technical in your analysis
- Focus on actionable intelligence, not generic advice
- Extract all IOCs mentioned in the article
- Confidence score should reflect the quality and specificity of the intelligence
- If no IOCs are mentioned, return an empty array
- Recommended actions should be prioritized (most critical first)
- Keep impact assessment concise but comprehensive`

// ArmorCTASystemPrompt defines the system context for Armor.com CTA generation
const ArmorCTASystemPrompt = `You are a marketing specialist for Armor.com, a cybersecurity services company specializing in:
- Managed Detection and Response (MDR)
- Security Operations Center (SOC) services
- Cloud security
- Compliance and risk management
- Incident response
- Threat intelligence

Your role is to analyze cybersecurity articles and generate relevant calls-to-action (CTAs) that match Armor's service offerings to the threats discussed.

You must respond ONLY with valid JSON in the following format:
{
  "type": "product|service|consultation",
  "title": "string",
  "url": "string"
}

CTA Types:
- product: Link to a specific Armor product page
- service: Link to a managed service offering
- consultation: Link to schedule a security consultation

Guidelines:
- Match the CTA to the specific threat or vulnerability discussed
- Use compelling, action-oriented titles
- URLs should use the pattern: https://armor.com/[relevant-path]
- Focus on how Armor can help mitigate the specific threat
- Keep titles concise (under 60 characters)
- Only recommend services that are truly relevant to the article content`

// BuildThreatAnalysisPrompt builds the user prompt for threat analysis
func BuildThreatAnalysisPrompt(title, content string, cves, vendors []string) string {
	var builder strings.Builder

	builder.WriteString("Analyze the following cybersecurity article and provide a comprehensive threat analysis:\n\n")

	builder.WriteString(fmt.Sprintf("**Title:** %s\n\n", title))

	if len(cves) > 0 {
		builder.WriteString(fmt.Sprintf("**CVEs Mentioned:** %s\n\n", strings.Join(cves, ", ")))
	}

	if len(vendors) > 0 {
		builder.WriteString(fmt.Sprintf("**Vendors/Products Affected:** %s\n\n", strings.Join(vendors, ", ")))
	}

	builder.WriteString("**Article Content:**\n")
	builder.WriteString(content)
	builder.WriteString("\n\n")

	builder.WriteString("Provide your analysis as JSON following the specified format. ")
	builder.WriteString("Extract all technical indicators (IPs, domains, hashes, URLs) mentioned. ")
	builder.WriteString("Focus on actionable intelligence that security teams can use immediately.")

	return builder.String()
}

// BuildArmorCTAPrompt builds the user prompt for Armor CTA generation
func BuildArmorCTAPrompt(title, content, threatType, attackVector string) string {
	var builder strings.Builder

	builder.WriteString("Generate a relevant call-to-action for Armor.com based on this cybersecurity article:\n\n")

	builder.WriteString(fmt.Sprintf("**Title:** %s\n\n", title))

	if threatType != "" {
		builder.WriteString(fmt.Sprintf("**Threat Type:** %s\n\n", threatType))
	}

	if attackVector != "" {
		builder.WriteString(fmt.Sprintf("**Attack Vector:** %s\n\n", attackVector))
	}

	builder.WriteString("**Article Content:**\n")
	builder.WriteString(content)
	builder.WriteString("\n\n")

	builder.WriteString("Provide a CTA as JSON that matches Armor's services to the threats discussed. ")
	builder.WriteString("Focus on how Armor can help organizations protect against or respond to this specific threat.")

	return builder.String()
}
