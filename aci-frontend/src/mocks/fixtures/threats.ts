/**
 * Mock Threat Fixtures
 * Mock data for threat intelligence endpoints with 50+ items for pagination testing
 */

import type {
  Threat,
  ThreatSummary,
  CVE,
  Severity,
  ExternalReference,
  Industry,
  Recommendation,
  DeepDive,
  MitreTechnique,
  IOC,
  TimelineEvent,
} from '../../types/threat';
import { ThreatCategory } from '../../types/threat';

// ============================================================================
// CVE Mock Data
// ============================================================================

const mockCVEs: readonly CVE[] = [
  {
    id: 'CVE-2024-1234',
    severity: 'critical',
    cvssScore: 9.8,
    description: 'Remote code execution vulnerability in Apache Struts',
  },
  {
    id: 'CVE-2024-5678',
    severity: 'critical',
    cvssScore: 9.1,
    description: 'SQL injection vulnerability in WordPress plugin',
  },
  {
    id: 'CVE-2024-9012',
    severity: 'high',
    cvssScore: 7.5,
    description: 'Buffer overflow vulnerability in OpenSSL',
  },
  {
    id: 'CVE-2024-3456',
    severity: 'high',
    cvssScore: 7.2,
    description: 'Privilege escalation in Linux kernel',
  },
  {
    id: 'CVE-2024-7890',
    severity: 'medium',
    cvssScore: 5.9,
    description: 'XSS vulnerability in React component library',
  },
  {
    id: 'CVE-2024-2468',
    severity: 'medium',
    cvssScore: 5.3,
    description: 'Information disclosure in Spring Framework',
  },
  {
    id: 'CVE-2024-1357',
    severity: 'low',
    cvssScore: 3.7,
    description: 'Minor path traversal in Node.js package',
  },
] as const;

// ============================================================================
// Industry and Reference Data
// ============================================================================

const INDUSTRIES: readonly Industry[] = [
  'finance',
  'healthcare',
  'manufacturing',
  'energy',
  'retail',
  'technology',
  'government',
  'education',
  'telecommunications',
  'transportation',
  'defense',
  'critical_infrastructure',
] as const;

const EXTERNAL_REFERENCES: readonly ExternalReference[] = [
  {
    title: 'MITRE ATT&CK Framework',
    url: 'https://attack.mitre.org',
    source: 'MITRE ATT&CK',
    type: 'mitre',
  },
  {
    title: 'CISA Known Exploited Vulnerabilities',
    url: 'https://www.cisa.gov/known-exploited-vulnerabilities',
    source: 'CISA',
    type: 'advisory',
  },
  {
    title: 'NVD CVE Database',
    url: 'https://nvd.nist.gov',
    source: 'NVD',
    type: 'cve',
  },
  {
    title: 'Microsoft Security Response Center Advisory',
    url: 'https://msrc.microsoft.com',
    source: 'Microsoft MSRC',
    type: 'advisory',
  },
  {
    title: 'Cybersecurity Threat Analysis Report',
    url: 'https://example.com/threat-analysis',
    source: 'ThreatPost',
    type: 'report',
  },
  {
    title: 'In-depth Technical Analysis',
    url: 'https://example.com/technical-deep-dive',
    source: 'SecurityWeek',
    type: 'article',
  },
] as const;

const MITRE_TECHNIQUES: readonly MitreTechnique[] = [
  {
    id: 'T1566.001',
    name: 'Spearphishing Attachment',
    tactic: 'Initial Access',
    url: 'https://attack.mitre.org/techniques/T1566/001',
  },
  {
    id: 'T1059.001',
    name: 'PowerShell',
    tactic: 'Execution',
    url: 'https://attack.mitre.org/techniques/T1059/001',
  },
  {
    id: 'T1547.001',
    name: 'Registry Run Keys / Startup Folder',
    tactic: 'Persistence',
    url: 'https://attack.mitre.org/techniques/T1547/001',
  },
  {
    id: 'T1055',
    name: 'Process Injection',
    tactic: 'Privilege Escalation',
    url: 'https://attack.mitre.org/techniques/T1055',
  },
  {
    id: 'T1027',
    name: 'Obfuscated Files or Information',
    tactic: 'Defense Evasion',
    url: 'https://attack.mitre.org/techniques/T1027',
  },
  {
    id: 'T1003.001',
    name: 'LSASS Memory',
    tactic: 'Credential Access',
    url: 'https://attack.mitre.org/techniques/T1003/001',
  },
  {
    id: 'T1083',
    name: 'File and Directory Discovery',
    tactic: 'Discovery',
    url: 'https://attack.mitre.org/techniques/T1083',
  },
  {
    id: 'T1021.001',
    name: 'Remote Desktop Protocol',
    tactic: 'Lateral Movement',
    url: 'https://attack.mitre.org/techniques/T1021/001',
  },
  {
    id: 'T1005',
    name: 'Data from Local System',
    tactic: 'Collection',
    url: 'https://attack.mitre.org/techniques/T1005',
  },
  {
    id: 'T1041',
    name: 'Exfiltration Over C2 Channel',
    tactic: 'Exfiltration',
    url: 'https://attack.mitre.org/techniques/T1041',
  },
  {
    id: 'T1486',
    name: 'Data Encrypted for Impact',
    tactic: 'Impact',
    url: 'https://attack.mitre.org/techniques/T1486',
  },
] as const;

// ============================================================================
// Threat Data Generator
// ============================================================================

const SOURCES = ['NVD', 'CISA', 'CVE Details', 'Proofpoint', 'SecurityWeek', 'BleepingComputer', 'ThreatPost', 'ZDI'] as const;

/**
 * Generate recommendations based on threat category
 */
function generateRecommendations(category: ThreatCategory): readonly Recommendation[] {
  const baseRecommendations: Record<ThreatCategory, Recommendation[]> = {
    [ThreatCategory.VULNERABILITY]: [
      {
        id: 'rec-patch-001',
        title: 'Apply Security Patches Immediately',
        description: 'Install the latest security updates from the vendor to address the vulnerability.',
        priority: 'immediate',
        category: 'patch',
      },
      {
        id: 'rec-monitor-001',
        title: 'Enable Vulnerability Scanning',
        description: 'Deploy automated scanning tools to detect similar vulnerabilities across your infrastructure.',
        priority: 'short_term',
        category: 'monitoring',
      },
      {
        id: 'rec-config-001',
        title: 'Harden System Configuration',
        description: 'Review and apply CIS benchmarks and security hardening guidelines.',
        priority: 'long_term',
        category: 'configuration',
      },
    ],
    [ThreatCategory.RANSOMWARE]: [
      {
        id: 'rec-backup-001',
        title: 'Verify Backup Integrity',
        description: 'Test backup restoration procedures and ensure backups are offline/immutable.',
        priority: 'immediate',
        category: 'backup',
      },
      {
        id: 'rec-access-001',
        title: 'Implement MFA on All Accounts',
        description: 'Enable multi-factor authentication for all user and admin accounts.',
        priority: 'immediate',
        category: 'access_control',
      },
      {
        id: 'rec-monitor-002',
        title: 'Deploy EDR Solutions',
        description: 'Implement endpoint detection and response tools to detect ransomware behavior.',
        priority: 'short_term',
        category: 'monitoring',
      },
    ],
    [ThreatCategory.PHISHING]: [
      {
        id: 'rec-training-001',
        title: 'Conduct Security Awareness Training',
        description: 'Train users to identify phishing emails and report suspicious messages.',
        priority: 'immediate',
        category: 'training',
      },
      {
        id: 'rec-config-002',
        title: 'Configure Email Security Controls',
        description: 'Enable SPF, DKIM, and DMARC to prevent email spoofing.',
        priority: 'short_term',
        category: 'configuration',
      },
      {
        id: 'rec-access-002',
        title: 'Enforce Password Policies',
        description: 'Implement strong password requirements and regular rotation schedules.',
        priority: 'long_term',
        category: 'access_control',
      },
    ],
    [ThreatCategory.APT]: [
      {
        id: 'rec-monitor-003',
        title: 'Enable Advanced Threat Detection',
        description: 'Deploy network traffic analysis and behavioral analytics tools.',
        priority: 'immediate',
        category: 'monitoring',
      },
      {
        id: 'rec-access-003',
        title: 'Implement Zero Trust Architecture',
        description: 'Deploy micro-segmentation and least-privilege access controls.',
        priority: 'short_term',
        category: 'access_control',
      },
      {
        id: 'rec-other-001',
        title: 'Conduct Threat Hunting Exercises',
        description: 'Proactively search for indicators of compromise in your environment.',
        priority: 'long_term',
        category: 'other',
      },
    ],
    [ThreatCategory.DATA_BREACH]: [
      {
        id: 'rec-access-004',
        title: 'Reset Compromised Credentials',
        description: 'Force password resets for all potentially affected accounts.',
        priority: 'immediate',
        category: 'access_control',
      },
      {
        id: 'rec-monitor-004',
        title: 'Monitor for Credential Abuse',
        description: 'Watch for unusual login patterns and access to sensitive data.',
        priority: 'short_term',
        category: 'monitoring',
      },
      {
        id: 'rec-config-003',
        title: 'Encrypt Sensitive Data',
        description: 'Implement encryption at rest and in transit for all sensitive information.',
        priority: 'long_term',
        category: 'configuration',
      },
    ],
    [ThreatCategory.SUPPLY_CHAIN]: [
      {
        id: 'rec-patch-002',
        title: 'Update Affected Dependencies',
        description: 'Remove compromised packages and update to verified clean versions.',
        priority: 'immediate',
        category: 'patch',
      },
      {
        id: 'rec-monitor-005',
        title: 'Implement SCA Tools',
        description: 'Deploy software composition analysis to detect vulnerable dependencies.',
        priority: 'short_term',
        category: 'monitoring',
      },
      {
        id: 'rec-config-004',
        title: 'Verify Package Integrity',
        description: 'Use package signing and checksum verification for all dependencies.',
        priority: 'long_term',
        category: 'configuration',
      },
    ],
    [ThreatCategory.ZERO_DAY]: [
      {
        id: 'rec-patch-003',
        title: 'Apply Emergency Patches',
        description: 'Deploy vendor-provided emergency patches or workarounds immediately.',
        priority: 'immediate',
        category: 'patch',
      },
      {
        id: 'rec-config-005',
        title: 'Implement Compensating Controls',
        description: 'Apply temporary mitigations until patches are available.',
        priority: 'immediate',
        category: 'configuration',
      },
      {
        id: 'rec-monitor-006',
        title: 'Monitor Exploitation Attempts',
        description: 'Deploy IDS/IPS rules to detect exploitation attempts.',
        priority: 'short_term',
        category: 'monitoring',
      },
    ],
    [ThreatCategory.DDOS]: [
      {
        id: 'rec-config-006',
        title: 'Enable DDoS Protection',
        description: 'Activate cloud-based DDoS mitigation services.',
        priority: 'immediate',
        category: 'configuration',
      },
      {
        id: 'rec-monitor-007',
        title: 'Monitor Traffic Patterns',
        description: 'Deploy traffic analysis tools to detect anomalous patterns.',
        priority: 'short_term',
        category: 'monitoring',
      },
      {
        id: 'rec-other-002',
        title: 'Review Incident Response Plan',
        description: 'Update DDoS response procedures and contact information.',
        priority: 'long_term',
        category: 'other',
      },
    ],
    [ThreatCategory.INSIDER_THREAT]: [
      {
        id: 'rec-access-005',
        title: 'Review Access Privileges',
        description: 'Audit user permissions and revoke unnecessary access.',
        priority: 'immediate',
        category: 'access_control',
      },
      {
        id: 'rec-monitor-008',
        title: 'Enable User Activity Monitoring',
        description: 'Deploy user and entity behavior analytics (UEBA) tools.',
        priority: 'short_term',
        category: 'monitoring',
      },
      {
        id: 'rec-training-002',
        title: 'Reinforce Security Policies',
        description: 'Conduct training on data handling and acceptable use policies.',
        priority: 'long_term',
        category: 'training',
      },
    ],
    [ThreatCategory.MALWARE]: [
      {
        id: 'rec-patch-004',
        title: 'Update Antivirus Signatures',
        description: 'Ensure all endpoint protection is updated with latest signatures.',
        priority: 'immediate',
        category: 'patch',
      },
      {
        id: 'rec-monitor-009',
        title: 'Scan All Systems',
        description: 'Perform full system scans to detect and remove malware.',
        priority: 'short_term',
        category: 'monitoring',
      },
      {
        id: 'rec-config-007',
        title: 'Restrict Execution Policies',
        description: 'Implement application whitelisting and restrict script execution.',
        priority: 'long_term',
        category: 'configuration',
      },
    ],
  };

  return baseRecommendations[category] || baseRecommendations[ThreatCategory.VULNERABILITY];
}

/**
 * Generate IOCs based on threat category
 */
function generateIOCs(category: ThreatCategory): readonly IOC[] {
  const iocSets: Record<ThreatCategory, IOC[]> = {
    [ThreatCategory.MALWARE]: [
      {
        type: 'hash_sha256',
        value: 'a'.repeat(64),
        description: 'Malicious payload hash',
        confidence: 'high',
        firstSeen: new Date(Date.now() - 86400000 * 7).toISOString(),
        lastSeen: new Date(Date.now() - 86400000).toISOString(),
      },
      {
        type: 'ip',
        value: '192.0.2.100',
        description: 'Command and control server',
        confidence: 'high',
        firstSeen: new Date(Date.now() - 86400000 * 5).toISOString(),
      },
      {
        type: 'domain',
        value: 'malicious-domain.example.com',
        description: 'C2 domain',
        confidence: 'medium',
      },
    ],
    [ThreatCategory.PHISHING]: [
      {
        type: 'email',
        value: 'phishing@example.com',
        description: 'Phishing sender address',
        confidence: 'high',
      },
      {
        type: 'url',
        value: 'https://fake-login.example.com',
        description: 'Credential harvesting page',
        confidence: 'high',
      },
      {
        type: 'domain',
        value: 'fake-domain.example.com',
        description: 'Phishing infrastructure',
        confidence: 'medium',
      },
    ],
    [ThreatCategory.RANSOMWARE]: [
      {
        type: 'hash_sha256',
        value: 'b'.repeat(64),
        description: 'Ransomware binary',
        confidence: 'high',
      },
      {
        type: 'ip',
        value: '198.51.100.50',
        description: 'Ransomware C2 server',
        confidence: 'high',
      },
      {
        type: 'file_name',
        value: 'README.txt',
        description: 'Ransom note filename',
        confidence: 'medium',
      },
    ],
    [ThreatCategory.APT]: [
      {
        type: 'ip',
        value: '203.0.113.75',
        description: 'APT infrastructure',
        confidence: 'high',
      },
      {
        type: 'domain',
        value: 'apt-c2.example.com',
        description: 'Command and control domain',
        confidence: 'high',
      },
      {
        type: 'hash_sha256',
        value: 'c'.repeat(64),
        description: 'Custom malware sample',
        confidence: 'high',
      },
    ],
    [ThreatCategory.VULNERABILITY]: [],
    [ThreatCategory.DATA_BREACH]: [],
    [ThreatCategory.SUPPLY_CHAIN]: [],
    [ThreatCategory.ZERO_DAY]: [],
    [ThreatCategory.DDOS]: [
      {
        type: 'ip',
        value: '192.0.2.0/24',
        description: 'DDoS botnet subnet',
        confidence: 'medium',
      },
    ],
    [ThreatCategory.INSIDER_THREAT]: [],
  };

  return iocSets[category] || [];
}

/**
 * Generate timeline events based on threat category
 */
function generateTimeline(): readonly TimelineEvent[] {
  const now = Date.now();
  const baseTimeline: TimelineEvent[] = [
    {
      timestamp: new Date(now - 86400000 * 30).toISOString(),
      event: 'Initial compromise detected',
      phase: 'initial_access',
      details: 'Attackers gained initial access through compromised credentials',
    },
    {
      timestamp: new Date(now - 86400000 * 28).toISOString(),
      event: 'Malware execution',
      phase: 'execution',
      details: 'Malicious payload executed on target system',
    },
    {
      timestamp: new Date(now - 86400000 * 27).toISOString(),
      event: 'Persistence mechanism established',
      phase: 'persistence',
      details: 'Registry keys modified for persistence',
    },
    {
      timestamp: new Date(now - 86400000 * 25).toISOString(),
      event: 'Privilege escalation',
      phase: 'privilege_escalation',
      details: 'Local admin rights obtained through vulnerability exploitation',
    },
    {
      timestamp: new Date(now - 86400000 * 23).toISOString(),
      event: 'Defense evasion',
      phase: 'defense_evasion',
      details: 'Antivirus disabled, logs cleared',
    },
    {
      timestamp: new Date(now - 86400000 * 20).toISOString(),
      event: 'Credential harvesting',
      phase: 'credential_access',
      details: 'LSASS memory dumped, credentials extracted',
    },
    {
      timestamp: new Date(now - 86400000 * 18).toISOString(),
      event: 'Network reconnaissance',
      phase: 'discovery',
      details: 'Active directory enumeration, network mapping',
    },
    {
      timestamp: new Date(now - 86400000 * 15).toISOString(),
      event: 'Lateral movement',
      phase: 'lateral_movement',
      details: 'Compromised additional systems using stolen credentials',
    },
    {
      timestamp: new Date(now - 86400000 * 10).toISOString(),
      event: 'Data collection',
      phase: 'collection',
      details: 'Sensitive files identified and staged for exfiltration',
    },
    {
      timestamp: new Date(now - 86400000 * 5).toISOString(),
      event: 'Data exfiltration',
      phase: 'exfiltration',
      details: 'Data transferred to external server via encrypted channel',
    },
    {
      timestamp: new Date(now - 86400000 * 2).toISOString(),
      event: 'Impact',
      phase: 'impact',
      details: 'Files encrypted, ransom note deployed',
    },
  ];

  return baseTimeline;
}

/**
 * Generate deep dive content for threats
 */
function generateDeepDive(category: ThreatCategory, index: number, isPremium: boolean): DeepDive {
  const isLocked = !isPremium;

  const detailedRemediationContent = `
## Step-by-Step Remediation

### Immediate Actions (0-24 hours)
1. **Isolate Affected Systems**
   - Disconnect compromised systems from the network
   - Preserve logs and evidence for forensic analysis
   - Document current system state

2. **Apply Emergency Patches**
   - Download security patches from vendor
   - Test patches in staging environment
   - Deploy to production systems using change management process

3. **Reset Credentials**
   - Force password resets for all potentially compromised accounts
   - Revoke and reissue API keys and service credentials
   - Review and update privileged access

### Short-Term Actions (1-7 days)
1. **Forensic Analysis**
   - Collect and analyze logs from affected systems
   - Identify indicators of compromise
   - Determine scope and timeline of breach

2. **System Hardening**
   - Apply CIS benchmarks
   - Disable unnecessary services
   - Implement least-privilege access

3. **Enhanced Monitoring**
   - Deploy EDR solutions
   - Configure SIEM alerts for IOCs
   - Implement network segmentation

### Long-Term Actions (1-3 months)
1. **Security Program Review**
   - Conduct lessons learned session
   - Update incident response procedures
   - Review and update security policies

2. **Preventive Controls**
   - Implement vulnerability management program
   - Deploy automated patch management
   - Conduct regular security assessments

3. **Training and Awareness**
   - Security awareness training for staff
   - Tabletop exercises for incident response team
   - Regular security briefings for leadership
  `.trim();

  const executiveSummaryContent = `
## Executive Summary

### Threat Overview
A critical security vulnerability has been identified that poses significant risk to organizational assets. This threat has been actively exploited in the wild and requires immediate attention.

### Business Impact
- **Financial Risk**: Potential data breach could result in $2M-$10M in costs
- **Operational Impact**: System downtime of 24-72 hours for remediation
- **Reputational Risk**: Customer trust and brand damage if exploited
- **Regulatory**: Potential compliance violations under GDPR/CCPA

### Recommended Actions
1. Allocate emergency resources for immediate patching (24 hours)
2. Authorize overtime for security team response coordination
3. Prepare customer communication if breach is detected
4. Budget for enhanced security controls ($50K-$200K)

### Timeline
- Immediate: 24 hours for critical patching
- Short-term: 7 days for comprehensive security review
- Long-term: 90 days for process improvements
  `.trim();

  return {
    isAvailable: true,
    isLocked,
    preview: isLocked
      ? 'Unlock premium content to access detailed technical analysis, MITRE ATT&CK mappings, indicators of compromise, and step-by-step remediation guidance.'
      : undefined,
    mitreTechniques: isPremium ? MITRE_TECHNIQUES.slice(0, 5) : [],
    iocs: isPremium ? generateIOCs(category) : [],
    timeline: isPremium ? generateTimeline() : [],
    detailedRemediation: isPremium ? detailedRemediationContent : '',
    executiveSummary: isPremium ? executiveSummaryContent : '',
    relatedThreats: isPremium ? [`threat-${String(((index + 5) % 60) + 1).padStart(3, '0')}`, `threat-${String(((index + 10) % 60) + 1).padStart(3, '0')}`] : [],
    ...(isPremium && {
      technicalAnalysis: {
        attackVector: 'Network-based exploitation via exposed web service',
        exploitationMethod: 'Remote code execution through deserialization vulnerability',
        affectedSystems: ['Apache Struts 2.x', 'Java-based web applications', 'Linux/Windows servers'],
        prerequisites: ['Exposed web service', 'Vulnerable Struts version', 'Network accessibility'],
        detectionMethods: [
          'Monitor for suspicious HTTP requests with serialized Java objects',
          'Detect abnormal process execution from web server processes',
          'Network IDS signatures for known exploit patterns',
          'File integrity monitoring for web application directories',
        ],
      },
      threatActorProfile: `
## Threat Actor Profile

### Attribution
This activity is attributed to APT29 (Cozy Bear), a sophisticated threat group believed to be sponsored by the Russian government.

### Historical Activity
- Active since at least 2008
- Known for targeting government, military, and critical infrastructure
- Previous campaigns include SolarWinds compromise (2020)

### Tactics and Techniques
- Highly sophisticated social engineering
- Custom malware development
- Long-term persistence and stealth
- Focus on intelligence gathering

### Motivation
Primary motivation appears to be cyber espionage and intelligence collection in support of national security objectives.
      `.trim(),
    }),
  };
}

const THREAT_TEMPLATES = [
  {
    titleTemplate: 'Critical: Zero-day RCE in {tech}',
    summaryTemplate: 'Remote code execution vulnerability discovered in {tech}',
    contentTemplate: '# Critical Remote Code Execution Vulnerability\n\nA critical remote code execution vulnerability has been discovered in {tech}. This vulnerability allows unauthenticated attackers to execute arbitrary code on affected systems.\n\n## Impact\n- Remote code execution\n- Full system compromise\n- Data exfiltration risk\n\n## Mitigation\nApply security patches immediately.',
    severity: 'critical' as Severity,
    category: ThreatCategory.VULNERABILITY,
    tech: ['Apache Struts', 'Nginx', 'PostgreSQL', 'Redis', 'Jenkins', 'GitLab', 'Docker', 'Kubernetes'],
  },
  {
    titleTemplate: 'High: Ransomware Campaign Targeting {industry}',
    summaryTemplate: '{malware} operators targeting {industry} organizations',
    contentTemplate: '# Ransomware Campaign Alert\n\n{malware} ransomware group has been observed targeting {industry} sector organizations. Multiple victims reported in the past 48 hours.\n\n## Tactics\n- Phishing emails with malicious attachments\n- Exploitation of known vulnerabilities\n- Lateral movement using stolen credentials\n\n## Recommendations\n- Implement email filtering\n- Patch known vulnerabilities\n- Enable MFA on all accounts',
    severity: 'high' as Severity,
    category: ThreatCategory.RANSOMWARE,
    industry: ['Finance', 'Healthcare', 'Manufacturing', 'Energy', 'Education', 'Retail'],
    malware: ['LockBit', 'BlackCat', 'Royal', 'Play', 'Cl0p'],
  },
  {
    titleTemplate: 'Medium: Phishing Campaign Using {method}',
    summaryTemplate: 'Credential harvesting campaign targeting {target}',
    contentTemplate: '# Phishing Campaign Detected\n\nA sophisticated phishing campaign has been detected using {method}. Attackers are targeting {target} to harvest credentials.\n\n## Indicators\n- Spoofed sender domains\n- Fake login pages\n- Social engineering tactics\n\n## Defense\n- User awareness training\n- Email authentication (SPF, DKIM, DMARC)\n- Multi-factor authentication',
    severity: 'medium' as Severity,
    category: ThreatCategory.PHISHING,
    method: ['fake MFA prompts', 'QR codes', 'legitimate-looking domains', 'OAuth consent abuse'],
    target: ['Microsoft 365 users', 'Google Workspace users', 'corporate executives', 'IT administrators'],
  },
  {
    titleTemplate: 'High: APT{number} {activity} Campaign',
    summaryTemplate: 'Advanced persistent threat activity targeting {sector}',
    contentTemplate: '# APT Campaign Analysis\n\nAPT{number} group has been observed conducting {activity} operations against {sector} targets.\n\n## TTPs\n- Spear-phishing with targeted lures\n- Custom malware deployment\n- Command and control infrastructure\n- Data exfiltration techniques\n\n## Attribution\nActivity consistent with known APT{number} patterns.',
    severity: 'high' as Severity,
    category: ThreatCategory.APT,
    number: ['29', '28', '41', '32', '1', '38'],
    activity: ['espionage', 'reconnaissance', 'data theft', 'surveillance'],
    sector: ['government', 'defense', 'technology', 'telecommunications'],
  },
  {
    titleTemplate: 'Medium: Supply Chain Compromise in {package}',
    summaryTemplate: 'Malicious code detected in popular {ecosystem} package',
    contentTemplate: '# Supply Chain Security Alert\n\nMalicious code has been discovered in {package}, a popular {ecosystem} package with significant download counts.\n\n## Details\n- Package version: X.X.X\n- Malicious behavior: credential theft\n- Affected downloads: 50k+\n\n## Response\n- Remove compromised package versions\n- Review dependency trees\n- Monitor for indicators of compromise',
    severity: 'medium' as Severity,
    category: ThreatCategory.SUPPLY_CHAIN,
    package: ['event-stream', 'ua-parser-js', 'node-ipc', 'colors', 'faker'],
    ecosystem: ['npm', 'PyPI', 'RubyGems', 'Maven'],
  },
  {
    titleTemplate: 'Critical: Zero-day Exploit for {product}',
    summaryTemplate: 'Actively exploited zero-day vulnerability discovered',
    contentTemplate: '# Zero-Day Vulnerability Alert\n\nA zero-day vulnerability is being actively exploited in {product}. Proof-of-concept code is publicly available.\n\n## Risk Level\nCRITICAL - Active exploitation observed\n\n## Immediate Actions\n- Apply emergency patches when available\n- Implement temporary workarounds\n- Monitor for exploitation attempts\n- Review logs for indicators of compromise',
    severity: 'critical' as Severity,
    category: ThreatCategory.ZERO_DAY,
    product: ['Windows', 'Chrome', 'Firefox', 'Safari', 'Exchange Server', 'VMware'],
  },
  {
    titleTemplate: 'High: Data Breach at {company}',
    summaryTemplate: '{records} records exposed in security incident',
    contentTemplate: '# Data Breach Notification\n\n{company} has disclosed a data breach affecting {records} user records.\n\n## Compromised Data\n- Email addresses\n- Hashed passwords\n- Personal information\n- Payment data (encrypted)\n\n## User Actions\n- Change passwords immediately\n- Enable MFA\n- Monitor for suspicious activity',
    severity: 'high' as Severity,
    category: ThreatCategory.DATA_BREACH,
    company: ['Major Tech Corp', 'Healthcare Provider', 'Financial Institution', 'Retail Chain'],
    records: ['500,000', '1.2 million', '2.5 million', '750,000'],
  },
  {
    titleTemplate: 'Medium: DDoS Campaign Against {target}',
    summaryTemplate: 'Distributed denial of service attacks observed',
    contentTemplate: '# DDoS Campaign Alert\n\nA coordinated DDoS campaign has been detected targeting {target} infrastructure.\n\n## Attack Vectors\n- Volumetric attacks (UDP/TCP floods)\n- Application layer attacks\n- DNS amplification\n\n## Mitigation\n- Rate limiting\n- Traffic filtering\n- CDN/DDoS protection services',
    severity: 'medium' as Severity,
    category: ThreatCategory.DDOS,
    target: ['financial services', 'government websites', 'gaming platforms', 'e-commerce sites'],
  },
  {
    titleTemplate: 'Low: Deprecated {library} Contains Vulnerabilities',
    summaryTemplate: 'Security issues in unmaintained software component',
    contentTemplate: '# Deprecated Software Alert\n\n{library} is no longer maintained and contains known security vulnerabilities.\n\n## Recommendations\n- Migrate to supported alternatives\n- Update dependency configurations\n- Review security advisories\n\n## Impact\nLow risk if not exposed to untrusted input',
    severity: 'low' as Severity,
    category: ThreatCategory.VULNERABILITY,
    library: ['jQuery 1.x', 'Angular.js', 'Prototype.js', 'Moment.js', 'Request'],
  },
];

/**
 * Generate mock threat with deterministic data based on index
 */
function generateThreat(index: number, full: boolean = false): Threat | ThreatSummary {
  const template = THREAT_TEMPLATES[index % THREAT_TEMPLATES.length];
  const sourceIdx = index % SOURCES.length;
  const daysAgo = Math.floor(index / 2);
  const publishedAt = new Date(Date.now() - daysAgo * 86400000);
  const createdAt = new Date(publishedAt.getTime() + 3600000);

  // Select random values from template arrays
  const getTechValue = (key: string): string => {
    const values = (template as Record<string, unknown>)[key];
    if (Array.isArray(values)) {
      return String(values[index % values.length]);
    }
    return '';
  };

  // Replace template placeholders
  let title: string = template.titleTemplate;
  let summary: string = template.summaryTemplate;
  let content: string = template.contentTemplate;

  const matches = title.match(/\{(\w+)\}/g);
  if (matches) {
    matches.forEach((match) => {
      const key = match.slice(1, -1);
      const value = getTechValue(key);
      title = title.replace(match, value);
      summary = summary.replace(match, value);
      content = content.replace(new RegExp(match, 'g'), value);
    });
  }

  // Determine if threat has CVEs
  const hasCVEs = template.category === ThreatCategory.VULNERABILITY && Math.random() > 0.3;
  const cveCount = hasCVEs ? Math.floor(Math.random() * 3) + 1 : 0;
  const cves = mockCVEs.slice(0, cveCount);

  const tags = [
    template.category,
    SOURCES[sourceIdx].toLowerCase(),
    template.severity,
  ];

  // Generate industries based on category
  const industryCount = Math.floor(Math.random() * 3) + 1; // 1-3 industries
  const industriesStart = index % (INDUSTRIES.length - industryCount);
  const industries = INDUSTRIES.slice(industriesStart, industriesStart + industryCount);

  // Determine if threat has deep dive content (80% have it available)
  const hasDeepDive = Math.random() > 0.2;

  const baseData = {
    id: `threat-${String(index + 1).padStart(3, '0')}`,
    title,
    summary,
    severity: template.severity,
    category: template.category,
    source: SOURCES[sourceIdx],
    publishedAt: publishedAt.toISOString(),
    cves: full ? cves : cves.map((c) => c.id),
    isBookmarked: Math.random() > 0.8,
    industries,
    hasDeepDive,
  };

  if (full) {
    // Generate external references (2-4 refs per threat)
    const refCount = Math.floor(Math.random() * 3) + 2;
    const refsStart = index % (EXTERNAL_REFERENCES.length - refCount + 1);
    const externalReferences = EXTERNAL_REFERENCES.slice(refsStart, refsStart + refCount);

    // Generate recommendations based on category
    const recommendations = generateRecommendations(template.category);

    // Simulate premium access (50% of users have premium)
    const isPremium = index % 2 === 0;

    return {
      ...baseData,
      content,
      sourceUrl: `https://example.com/threat-${index + 1}`,
      createdAt: createdAt.toISOString(),
      updatedAt: createdAt.toISOString(),
      cves,
      tags,
      viewCount: Math.floor(Math.random() * 1000) + 50,
      externalReferences,
      recommendations,
      ...(hasDeepDive && { deepDive: generateDeepDive(template.category, index, isPremium) }),
    } as Threat;
  }

  return baseData as ThreatSummary;
}

// ============================================================================
// Exported Mock Data
// ============================================================================

/**
 * Generate 60 mock threat summaries for pagination testing
 */
export const mockThreats: readonly ThreatSummary[] = Array.from({ length: 60 }, (_, i) =>
  generateThreat(i, false)
) as ThreatSummary[];

/**
 * Generate full threat details (lazy generation on demand)
 */
export function getMockThreatById(id: string): Threat | undefined {
  const match = id.match(/^threat-(\d+)$/);
  if (!match) {
    return undefined;
  }

  const index = parseInt(match[1], 10) - 1;
  if (index < 0 || index >= 60) {
    return undefined;
  }

  return generateThreat(index, true) as Threat;
}

/**
 * Filter threats based on query parameters
 */
export function filterThreats(params: {
  severity?: readonly Severity[];
  category?: readonly ThreatCategory[];
  source?: readonly string[];
  search?: string;
  startDate?: string;
  endDate?: string;
}): ThreatSummary[] {
  return mockThreats.filter((threat) => {
    // Severity filter
    if (params.severity && params.severity.length > 0) {
      if (!params.severity.includes(threat.severity)) {
        return false;
      }
    }

    // Category filter
    if (params.category && params.category.length > 0) {
      if (!params.category.includes(threat.category)) {
        return false;
      }
    }

    // Source filter
    if (params.source && params.source.length > 0) {
      if (!params.source.includes(threat.source)) {
        return false;
      }
    }

    // Search filter (case-insensitive)
    if (params.search) {
      const searchLower = params.search.toLowerCase();
      const matchesTitle = threat.title.toLowerCase().includes(searchLower);
      const matchesSummary = threat.summary.toLowerCase().includes(searchLower);
      if (!matchesTitle && !matchesSummary) {
        return false;
      }
    }

    // Date range filter
    if (params.startDate) {
      const threatDate = new Date(threat.publishedAt);
      const startDate = new Date(params.startDate);
      if (threatDate < startDate) {
        return false;
      }
    }

    if (params.endDate) {
      const threatDate = new Date(threat.publishedAt);
      const endDate = new Date(params.endDate);
      if (threatDate > endDate) {
        return false;
      }
    }

    return true;
  });
}
