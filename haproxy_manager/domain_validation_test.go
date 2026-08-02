package haproxymanager

import "testing"

func TestValidateDomainNameAcceptsRealHostnames(t *testing.T) {
	valid := []string{
		"example.com",
		"sub.example.com",
		"a.b.c.d.example.co.uk",
		"my-app-1.example.com",
		"localhost",
		"123.example.com",
	}
	for _, domain := range valid {
		if err := ValidateDomainName(domain); err != nil {
			t.Errorf("%q should be accepted, got %v", domain, err)
		}
	}
}

func TestValidateDomainNameRejectsAclInjection(t *testing.T) {
	invalid := []string{
		"attacker.example } or { src 0.0.0.0/0 } #",
		"example.com }",
		"{ src 0.0.0.0/0 }",
		"example.com or { src 0.0.0.0/0 }",
		"example.com\n  acl anything",
		"example.com/../../etc/haproxy",
		"example.com:8080",
		"",
		"   ",
		"..",
		"example..com",
		"-example.com",
		"example-.com",
	}
	for _, domain := range invalid {
		if err := ValidateDomainName(domain); err == nil {
			t.Errorf("%q should be rejected", domain)
		}
	}
}

func TestValidateDomainNameRejectsOversizedInput(t *testing.T) {
	label := make([]byte, 64)
	for i := range label {
		label[i] = 'a'
	}
	if err := ValidateDomainName(string(label) + ".com"); err == nil {
		t.Error("a label longer than 63 characters should be rejected")
	}

	long := ""
	for range 40 {
		long += "abcdef."
	}
	if err := ValidateDomainName(long + "com"); err == nil {
		t.Error("a name longer than 253 characters should be rejected")
	}
}
