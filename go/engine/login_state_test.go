package engine

import (
	"errors"
	"runtime"
	"testing"

	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/protocol/go"
)

// TODO: These tests should really be in libkb/. However, any test
// that creates new users have to remain in engine/ for now. Fix this.

// Test that LoginState and Session are in sync regarding whether a
// user is logged in.
func TestLoginLogout(t *testing.T) {
	tc := SetupEngineTest(t, "login logout")
	defer tc.Cleanup()

	if err := tc.G.LoginState().AssertLoggedOut(); err != nil {
		t.Error("Unexpectedly logged in (Session)")
	}

	if tc.G.LoginState().IsLoggedIn() {
		t.Error("Unexpectedly logged in (LoginState)")
	}

	// Logging out when not logged in should still work.
	tc.G.LoginState().Logout()

	fu := CreateAndSignupFakeUser(tc, "login")

	if err := tc.G.LoginState().AssertLoggedIn(); err != nil {
		t.Error("Unexpectedly logged out (Session)")
	}

	tc.G.LoginState().Logout()

	if err := tc.G.LoginState().AssertLoggedOut(); err != nil {
		t.Error("Unexpectedly logged in (Session)")
	}

	if tc.G.LoginState().IsLoggedIn() {
		t.Error("Unexpectedly logged in (LoginState)")
	}

	// Logging out twice should still work.
	tc.G.LoginState().Logout()

	if err := tc.G.LoginState().AssertLoggedOut(); err != nil {
		t.Error("Unexpectedly logged in (Session)")
	}

	if tc.G.LoginState().IsLoggedIn() {
		t.Error("Unexpectedly logged in (LoginState)")
	}

	secretUI := &libkb.TestSecretUI{Passphrase: fu.Passphrase}
	if err := tc.G.LoginState().LoginWithPrompt("", nil, secretUI); err != nil {
		t.Error(err)
	}

	if err := tc.G.LoginState().AssertLoggedIn(); err != nil {
		t.Error("Unexpectedly logged out (Session)")
	}

	if !tc.G.LoginState().IsLoggedIn() {
		t.Error("Unexpectedly logged out (LoginState)")
	}
}

// This mock (and the similar ones below) may be used from a goroutine
// different from the main one, so don't mess with testing.T (which
// isn't safe to use from a non-main goroutine) directly, and instead
// have a LastErr field.
type GetSecretMock struct {
	Passphrase  string
	StoreSecret bool
	Called      bool
	LastErr     error
}

func (m *GetSecretMock) GetSecret(arg keybase1.SecretEntryArg, _ *keybase1.SecretEntryArg) (*keybase1.SecretEntryRes, error) {
	if m.Called {
		m.LastErr = errors.New("GetSecret unexpectedly called more than once")
		return nil, m.LastErr
	}
	m.Called = true
	storeSecret := arg.UseSecretStore && m.StoreSecret
	return &keybase1.SecretEntryRes{Text: m.Passphrase, StoreSecret: storeSecret}, nil
}

func (m *GetSecretMock) GetNewPassphrase(keybase1.GetNewPassphraseArg) (string, error) {
	m.LastErr = errors.New("GetSecret unexpectedly called more than once")
	return "invalid passphrase", m.LastErr
}

func (m *GetSecretMock) GetKeybasePassphrase(keybase1.GetKeybasePassphraseArg) (string, error) {
	m.LastErr = errors.New("GetKeybasePassphrase unexpectedly called")
	return "invalid passphrase", m.LastErr
}

func (m *GetSecretMock) CheckLastErr(t *testing.T) {
	if m.LastErr != nil {
		t.Fatal(m.LastErr)
	}
}

// Test that login works while already logged in.
func TestLoginWhileAlreadyLoggedIn(t *testing.T) {
	tc := SetupEngineTest(t, "login while already logged in")
	defer tc.Cleanup()

	// Logs the user in.
	fu := CreateAndSignupFakeUser(tc, "li")

	// These should all work, since the username matches.

	if err := tc.G.LoginState().LoginWithPrompt("", nil, nil); err != nil {
		t.Error(err)
	}

	if err := tc.G.LoginState().LoginWithPrompt(fu.Username, nil, nil); err != nil {
		t.Error(err)
	}

	if err := tc.G.LoginState().LoginWithStoredSecret(fu.Username); err != nil {
		t.Error(err)
	}

	if err := tc.G.LoginState().LoginWithPassphrase(fu.Username, "", false); err != nil {
		t.Error(err)
	}

	// This should fail.
	if _, ok := tc.G.LoginState().LoginWithPrompt("other", nil, nil).(libkb.LoggedInError); !ok {
		t.Fatal("Did not get expected LoggedIn error")
	}
}

// Test that login fails with a nonexistent user.
func TestLoginNonexistent(t *testing.T) {
	tc := SetupEngineTest(t, "login nonexistent")
	defer tc.Cleanup()

	_ = CreateAndSignupFakeUser(tc, "ln")

	tc.G.LoginState().Logout()

	err := tc.G.LoginState().LoginWithPrompt("nonexistent", nil, nil)
	if _, ok := err.(libkb.AppStatusError); !ok {
		t.Error("Did not get expected AppStatusError")
	}
}

// Test that the login prompts for a passphrase for the pubkey first.
func TestLoginWithPromptPubkey(t *testing.T) {
	tc := SetupEngineTest(t, "login with prompt (pubkey)")
	defer tc.Cleanup()

	fu := CreateAndSignupFakeUser(tc, "lwpp")

	tc.G.LoginState().Logout()

	mockGetSecret := &GetSecretMock{
		Passphrase: fu.Passphrase,
	}
	if err := tc.G.LoginState().LoginWithPrompt("", nil, mockGetSecret); err != nil {
		t.Error(err)
	}

	mockGetSecret.CheckLastErr(t)

	if !mockGetSecret.Called {
		t.Errorf("secretUI.GetSecret() unexpectedly not called")
	}

	tc.G.LoginState().Logout()

	mockGetSecret.Called = false
	if err := tc.G.LoginState().LoginWithPrompt(fu.Username, nil, mockGetSecret); err != nil {
		t.Error(err)
	}

	if !mockGetSecret.Called {
		t.Errorf("secretUI.GetSecret() unexpectedly not called")
	}

	// The interaction with the loginUI is covered by
	// TestLoginWithPromptPassphrase below.
}

type GetUsernameMock struct {
	Username string
	Called   bool
	LastErr  error
}

func (m *GetUsernameMock) GetEmailOrUsername(int) (string, error) {
	if m.Called {
		m.LastErr = errors.New("GetEmailOrUsername unexpectedly called more than once")
		return "invalid username", m.LastErr
	}
	m.Called = true
	return m.Username, nil
}

func (m *GetUsernameMock) CheckLastErr(t *testing.T) {
	if m.LastErr != nil {
		t.Fatal(m.LastErr)
	}
}

type GetKeybasePassphraseMock struct {
	Passphrase string
	Called     bool
	LastErr    error
}

func (m *GetKeybasePassphraseMock) GetSecret(keybase1.SecretEntryArg, *keybase1.SecretEntryArg) (*keybase1.SecretEntryRes, error) {
	return nil, errors.New("Fail pubkey login")
}

func (m *GetKeybasePassphraseMock) GetNewPassphrase(keybase1.GetNewPassphraseArg) (string, error) {
	m.LastErr = errors.New("GetNewPassphrase unexpectedly called")
	return "invalid passphrase", m.LastErr
}

func (m *GetKeybasePassphraseMock) GetKeybasePassphrase(keybase1.GetKeybasePassphraseArg) (string, error) {
	if m.Called {
		m.LastErr = errors.New("GetKeybasePassphrase unexpectedly called more than once")
		return "invalid passphrase", m.LastErr
	}
	m.Called = true
	return m.Passphrase, nil
}

func (m *GetKeybasePassphraseMock) CheckLastErr(t *testing.T) {
	if m.LastErr != nil {
		t.Fatal(m.LastErr)
	}
}

// Test that the login falls back to a passphrase login if pubkey
// login fails.
func TestLoginWithPromptPassphrase(t *testing.T) {
	tc := SetupEngineTest(t, "login with prompt (passphrase)")
	defer tc.Cleanup()

	fu := CreateAndSignupFakeUser(tc, "lwpp")

	tc.G.Logout()

	mockGetKeybasePassphrase := &GetKeybasePassphraseMock{
		Passphrase: fu.Passphrase,
	}
	if err := tc.G.LoginState().LoginWithPrompt("", nil, mockGetKeybasePassphrase); err != nil {
		t.Error(err)
	}

	if !mockGetKeybasePassphrase.Called {
		t.Errorf("secretUI.GetKeybasePassphrase() unexpectedly not called")
	}

	tc.G.Logout()

	mockGetKeybasePassphrase.Called = false
	if err := tc.G.LoginState().LoginWithPrompt(fu.Username, nil, mockGetKeybasePassphrase); err != nil {
		t.Error(err)
	}

	mockGetKeybasePassphrase.CheckLastErr(t)

	if !mockGetKeybasePassphrase.Called {
		t.Errorf("secretUI.GetKeybasePassphrase() unexpectedly not called")
	}

	tc.G.Logout()

	// Clear out the username stored in G.Env.
	// TODO: Figure out a cleaner way to do this.
	tc.G.Env = libkb.NewEnv(nil, nil)

	mockGetUsername := &GetUsernameMock{
		Username: fu.Username,
	}
	mockGetKeybasePassphrase.Called = false
	if err := tc.G.LoginState().LoginWithPrompt("", mockGetUsername, mockGetKeybasePassphrase); err != nil {
		t.Error(err)
	}

	mockGetUsername.CheckLastErr(t)
	mockGetKeybasePassphrase.CheckLastErr(t)

	if !mockGetUsername.Called {
		t.Errorf("loginUI.GetEmailOrUsername() unexpectedly not called")
	}

	if !mockGetKeybasePassphrase.Called {
		t.Errorf("secretUI.GetKeybasePassphrase() unexpectedly not called")
	}
}

func userHasStoredSecretViaConfiguredAccounts(tc *libkb.TestContext, username string) bool {
	configuredAccounts, err := tc.G.LoginState().GetConfiguredAccounts()
	if err != nil {
		tc.T.Error(err)
		return false
	}

	for _, configuredAccount := range configuredAccounts {
		if configuredAccount.Username == username {
			return configuredAccount.HasStoredSecret
		}
	}
	return false
}

func userHasStoredSecretViaSecretStore(tc *libkb.TestContext, username string) bool {
	secretStore := libkb.NewSecretStore(username)
	if secretStore == nil {
		tc.T.Errorf("SecretStore for %s unexpectedly nil", username)
		return false
	}
	_, err := secretStore.RetrieveSecret()
	// TODO: Have RetrieveSecret return platform-independent errors
	// so that we can make sure we got the right one.
	return (err == nil)
}

func userHasStoredSecret(tc *libkb.TestContext, username string) bool {
	hasStoredSecret1 := userHasStoredSecretViaConfiguredAccounts(tc, username)
	hasStoredSecret2 := userHasStoredSecretViaSecretStore(tc, username)
	if hasStoredSecret1 != hasStoredSecret2 {
		tc.T.Errorf("user %s has stored secret via configured accounts = %t, but via secret store = %t", username, hasStoredSecret1, hasStoredSecret2)
	}
	return hasStoredSecret1
}

// Test that the login flow using the secret store works.
func TestLoginWithStoredSecret(t *testing.T) {
	// TODO: Get this working on non-OS X platforms (by mocking
	// out the SecretStore).
	if !libkb.HasSecretStore() {
		t.Skip("Skipping test since there is no secret store")
	}

	tc := SetupEngineTest(t, "login with stored secret")
	defer tc.Cleanup()

	fu := CreateAndSignupFakeUser(tc, "lwss")
	tc.G.LoginState().Logout()

	if userHasStoredSecret(&tc, fu.Username) {
		t.Errorf("User %s unexpectedly has a stored secret", fu.Username)
	}

	mockGetSecret := &GetSecretMock{
		Passphrase:  fu.Passphrase,
		StoreSecret: true,
	}
	if err := tc.G.LoginState().LoginWithPrompt("", nil, mockGetSecret); err != nil {
		t.Error(err)
	}

	mockGetSecret.CheckLastErr(t)

	if !mockGetSecret.Called {
		t.Errorf("secretUI.GetSecret() unexpectedly not called")
	}

	tc.G.LoginState().Logout()

	if !userHasStoredSecret(&tc, fu.Username) {
		t.Errorf("User %s unexpectedly does not have a stored secret", fu.Username)
	}

	// TODO: Mock out the SecretStore and make sure that it's
	// actually consulted.
	if err := tc.G.LoginState().LoginWithStoredSecret(fu.Username); err != nil {
		t.Error(err)
	}

	tc.G.LoginState().Logout()

	tc.G.LoginState().ClearStoredSecret(fu.Username)

	if userHasStoredSecret(&tc, fu.Username) {
		t.Errorf("User %s unexpectedly has a stored secret", fu.Username)
	}

	if err := tc.G.LoginState().LoginWithStoredSecret(fu.Username); err == nil {
		t.Error("Did not get expected error")
	}

	if err := tc.G.LoginState().LoginWithStoredSecret(""); err == nil {
		t.Error("Did not get expected error")
	}

	fu = CreateAndSignupFakeUser(tc, "lwss")
	tc.G.LoginState().Logout()

	if err := tc.G.LoginState().LoginWithStoredSecret(fu.Username); err == nil {
		t.Error("Did not get expected error")
	}
}

// Test that the login flow with passphrase correctly denies bad
// usernames/passphrases.
func TestLoginWithPassphraseErrors(t *testing.T) {
	tc := SetupEngineTest(t, "login with passphrase (errors)")
	defer tc.Cleanup()

	fu := CreateAndSignupFakeUser(tc, "lwpe")
	tc.G.LoginState().Logout()

	err := tc.G.LoginState().LoginWithPassphrase("", "", false)
	if _, ok := err.(libkb.AppStatusError); !ok {
		t.Error("Did not get expected AppStatusError")
	}

	err = tc.G.LoginState().LoginWithPassphrase(fu.Username, "wrong passphrase", false)
	if _, ok := err.(libkb.PassphraseError); !ok {
		t.Error("Did not get expected PassphraseError")
	}
}

// Test that the login flow with passphrase but without saving the
// secret works.
func TestLoginWithPassphraseNoStore(t *testing.T) {
	// TODO: Get this working on non-OS X platforms (by mocking
	// out the SecretStore).
	if !libkb.HasSecretStore() {
		t.Skip("Skipping test since there is no secret store")
	}

	tc := SetupEngineTest(t, "login with passphrase (no store)")
	defer tc.Cleanup()

	fu := CreateAndSignupFakeUser(tc, "lwpns")
	tc.G.LoginState().Logout()

	if err := tc.G.LoginState().LoginWithPassphrase(fu.Username, fu.Passphrase, false); err != nil {
		t.Error(err)
	}

	tc.G.LoginState().Logout()

	if err := tc.G.LoginState().LoginWithStoredSecret(fu.Username); err == nil {
		t.Error("Did not get expected error")
	}

	if userHasStoredSecret(&tc, fu.Username) {
		t.Errorf("User %s unexpectedly has a stored secret", fu.Username)
	}
}

// Test that the login flow with passphrase and with saving the secret
// works.
func TestLoginWithPassphraseWithStore(t *testing.T) {
	// TODO: Get this working on non-OS X platforms (by mocking
	// out the SecretStore).
	if !libkb.HasSecretStore() {
		t.Skip("Skipping test since there is no secret store")
	}

	tc := SetupEngineTest(t, "login with passphrase (with store)")
	defer tc.Cleanup()

	fu := CreateAndSignupFakeUser(tc, "lwpws")
	tc.G.LoginState().Logout()

	if userHasStoredSecret(&tc, fu.Username) {
		t.Errorf("User %s unexpectedly has a stored secret", fu.Username)
	}

	if err := tc.G.LoginState().LoginWithPassphrase(fu.Username, fu.Passphrase, true); err != nil {
		t.Error(err)
	}

	tc.G.LoginState().Logout()

	if !userHasStoredSecret(&tc, fu.Username) {
		t.Errorf("User %s unexpectedly does not have a stored secret", fu.Username)
	}

	// TODO: Mock out the SecretStore and make sure that it's
	// actually consulted.
	if err := tc.G.LoginState().LoginWithStoredSecret(fu.Username); err != nil {
		t.Error(err)
	}

	tc.G.LoginState().ClearStoredSecret(fu.Username)

	if userHasStoredSecret(&tc, fu.Username) {
		t.Errorf("User %s unexpectedly has a stored secret", fu.Username)
	}
}

// TODO: Test LoginWithPassphrase with pubkey login failing.

func TestExternalFuncGoexit(t *testing.T) {
	tc := SetupEngineTest(t, "ExternalFunc goexit")
	defer tc.Cleanup()

	// This should not cause a hang, and an error should be
	// returned.
	err := tc.G.LoginState().ExternalFunc(func() error {
		runtime.Goexit()
		return nil
	})

	if err == nil {
		t.Error("Error unexpectedly nil")
	}
}
