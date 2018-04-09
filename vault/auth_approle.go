package vault

import (
	"git-platform.dbc.dk/platform/morph/nix"
	vault "github.com/hashicorp/vault/api"
)

func newSecretID(client *vault.Client, host nix.Host) (*vault.Secret, error) {

	r := client.NewRequest("POST", "/v1/auth/approle/role/"+host.TargetHost+"/secret-id")

	if err := r.SetJSONBody(secretIDCreateRequest{}); err != nil {
		return nil, err
	}

	resp, err := client.RawRequest(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return vault.ParseSecret(resp.Body)
}

func syncAppRole(client *vault.Client, host nix.Host) (*vault.Secret, error) {

	req1 := client.NewRequest("POST", "/v1/auth/approle/role/"+host.TargetHost)

	if err := req1.SetJSONBody(appRoleCreateRequest{

		BindSecretID:    true,
		BoundCIDRList:   host.Vault.CIDRs,
		Policies:        host.Vault.Policies,
		SecretIDNumUses: 0,
		SecretIDTTL:     host.Vault.TTL,
	}); err != nil {
		return nil, err
	}

	_, err := client.RawRequest(req1)
	if err != nil {
		return nil, err
	}

	// Annoyingly, Vault doesn't give us the role-id in the response from app-role create/update
	// we have to retrieve it with a subsequent GET-request
	req2 := client.NewRequest("GET", "/v1/auth/approle/role/"+host.TargetHost+"/role-id")
	resp, err := client.RawRequest(req2)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return vault.ParseSecret(resp.Body)
}

type secretIDCreateRequest struct {
	Metadata map[string]string `json:"meta,omitempty"`
}

type appRoleCreateRequest struct {
	BindSecretID    bool     `json:"bind_secret_id,omitempty"`
	BoundCIDRList   []string `json:"bound_cidr_list,omitempty"`
	Policies        []string `json:"policies,omitempty"`
	SecretIDNumUses int      `json:"secret_id_num_uses,omitempty"`
	SecretIDTTL     string   `json:"secret_id_ttl,omitempty"`
}
