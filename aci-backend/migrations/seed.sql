-- Seed Data for ACI Backend
-- Description: Initial data for categories and sources
-- Author: Database Developer Agent
-- Date: 2025-12-11
-- Note: This should be run AFTER all migrations are complete

-- Seed Categories (8 cybersecurity categories)
INSERT INTO categories (name, slug, description, color, icon) VALUES
    ('Vulnerabilities', 'vulnerabilities', 'Security vulnerabilities, CVEs, and patches', '#ef4444', 'shield-exclamation'),
    ('Ransomware', 'ransomware', 'Ransomware attacks, campaigns, and threat actors', '#dc2626', 'lock-closed'),
    ('Data Breaches', 'data-breaches', 'Data breaches, leaks, and exposures', '#f97316', 'database'),
    ('Threat Actors', 'threat-actors', 'APT groups, cybercriminal organizations, and attribution', '#8b5cf6', 'user-group'),
    ('Malware', 'malware', 'Malware analysis, campaigns, and indicators', '#ec4899', 'bug'),
    ('Phishing', 'phishing', 'Phishing campaigns, business email compromise, and social engineering', '#f59e0b', 'mail'),
    ('Compliance', 'compliance', 'Regulatory compliance, standards, and frameworks', '#10b981', 'clipboard-check'),
    ('Industry News', 'industry-news', 'Cybersecurity industry news, acquisitions, and trends', '#3b82f6', 'newspaper')
ON CONFLICT (slug) DO NOTHING;

-- Seed Sources (10 trusted cybersecurity news sources)
INSERT INTO sources (name, url, description, is_active, trust_score) VALUES
    (
        'CISA',
        'https://www.cisa.gov/news-events/cybersecurity-advisories',
        'Cybersecurity and Infrastructure Security Agency - Official US government cybersecurity advisories',
        true,
        0.95
    ),
    (
        'National Vulnerability Database',
        'https://nvd.nist.gov/vuln/search',
        'NIST National Vulnerability Database - Comprehensive CVE database',
        true,
        0.98
    ),
    (
        'Krebs on Security',
        'https://krebsonsecurity.com',
        'In-depth security news and investigation by Brian Krebs',
        true,
        0.90
    ),
    (
        'BleepingComputer',
        'https://www.bleepingcomputer.com',
        'Technology news and computer security information',
        true,
        0.85
    ),
    (
        'The Hacker News',
        'https://thehackernews.com',
        'Cybersecurity news and analysis',
        true,
        0.82
    ),
    (
        'Dark Reading',
        'https://www.darkreading.com',
        'Cybersecurity news, analysis, and research',
        true,
        0.88
    ),
    (
        'Threatpost',
        'https://threatpost.com',
        'Breaking cybersecurity news and threat analysis',
        true,
        0.85
    ),
    (
        'SecurityWeek',
        'https://www.securityweek.com',
        'Enterprise security news and analysis',
        true,
        0.87
    ),
    (
        'US-CERT',
        'https://www.cisa.gov/uscert',
        'United States Computer Emergency Readiness Team alerts',
        true,
        0.95
    ),
    (
        'MITRE ATT&CK',
        'https://attack.mitre.org',
        'MITRE ATT&CK framework - Adversary tactics and techniques',
        true,
        0.98
    )
ON CONFLICT (url) DO NOTHING;

-- Verify seed data
DO $$
DECLARE
    category_count INTEGER;
    source_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO category_count FROM categories;
    SELECT COUNT(*) INTO source_count FROM sources;

    RAISE NOTICE 'Seed data loaded successfully:';
    RAISE NOTICE '  - Categories: %', category_count;
    RAISE NOTICE '  - Sources: %', source_count;

    IF category_count < 8 THEN
        RAISE WARNING 'Expected 8 categories, found %', category_count;
    END IF;

    IF source_count < 10 THEN
        RAISE WARNING 'Expected 10 sources, found %', source_count;
    END IF;
END $$;
