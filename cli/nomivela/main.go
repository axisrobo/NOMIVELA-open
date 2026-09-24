// Command nomivela is a thin command-line client for the NOMIVELA Agent
// Registry API. It prints JSON so its output can be piped to other tools.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	nomivela "github.com/axisrobo/nomivela-open/sdk/go"
)

var stdout io.Writer = os.Stdout

type config struct {
	api         string
	actor       string
	reason      string
	evidenceRef string
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: nomivela <namespace|agent|identity|workload|instance|evidence|events> <action> [flags]")
	}

	switch args[0] {
	case "namespace", "namespaces":
		return runNamespace(ctx, args[1:])
	case "agent", "agents":
		return runAgent(ctx, args[1:])
	case "identity", "identities":
		return runIdentity(ctx, args[1:])
	case "workload", "workloads":
		return runWorkload(ctx, args[1:])
	case "instance", "instances":
		return runInstance(ctx, args[1:])
	case "evidence":
		return runEvidence(ctx, args[1:])
	case "events":
		return runEvents(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func newConfig() config {
	return config{
		api:   envOr("NOMIVELA_API_URL", "http://localhost:8080"),
		actor: envOr("NOMIVELA_ACTOR", "nomivela-cli"),
	}
}

// globalFlags registers flags shared by every command.
func (c *config) globalFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.api, "api", c.api, "API base URL")
	fs.StringVar(&c.actor, "actor", c.actor, "actor recorded for mutations")
	fs.StringVar(&c.reason, "reason", "", "reason recorded for mutations")
	fs.StringVar(&c.evidenceRef, "evidence-ref", "", "evidence reference recorded for mutations")
}

func (c config) client() *nomivela.Client {
	return nomivela.New(c.api, nomivela.WithActor(c.actor))
}

func (c config) mutation() nomivela.Mutation {
	return nomivela.Mutation{Reason: c.reason, EvidenceRef: c.evidenceRef}
}

func runNamespace(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: nomivela namespace <create|list> [flags]")
	}
	c := newConfig()

	switch args[0] {
	case "create":
		fs := flag.NewFlagSet("namespace create", flag.ContinueOnError)
		namespace := fs.String("namespace", "", "canonical authority namespace")
		root := fs.String("root-ref", "", "authority root reference")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().CreateNamespace(ctx, *namespace, *root, c.mutation())
		if err != nil {
			return err
		}
		return print(result)

	case "list":
		fs := flag.NewFlagSet("namespace list", flag.ContinueOnError)
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().ListNamespaces(ctx)
		if err != nil {
			return err
		}
		return print(result)

	default:
		return fmt.Errorf("unknown namespace action %q", args[0])
	}
}

func runAgent(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: nomivela agent <create|list|activate|suspend|contain> [flags]")
	}
	c := newConfig()

	switch args[0] {
	case "create":
		fs := flag.NewFlagSet("agent create", flag.ContinueOnError)
		ref := fs.String("ref", "", "agent reference")
		name := fs.String("name", "", "display name")
		purpose := fs.String("purpose", "", "purpose")
		sponsor := fs.String("sponsor", "", "sponsor reference")
		owner := fs.String("owner", "", "owner reference")
		risk := fs.String("risk", "medium", "risk class")
		class := fs.String("class", "", "agent class: embedded, organizational, user, asset_twin, personal_twin, service")
		carriers := fs.String("carrier-refs", "", "comma-separated carrier references")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().CreateAgent(ctx, nomivela.Agent{
			AgentRef: *ref, Name: *name, Purpose: *purpose, SponsorRef: *sponsor, OwnerRef: *owner, RiskClass: *risk,
			AgentClass: *class, CarrierRefs: splitList(*carriers),
		}, c.mutation())
		if err != nil {
			return err
		}
		return print(result)

	case "list":
		fs := flag.NewFlagSet("agent list", flag.ContinueOnError)
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().ListAgents(ctx)
		if err != nil {
			return err
		}
		return print(result)

	case "lifecycle":
		fs := flag.NewFlagSet("agent lifecycle", flag.ContinueOnError)
		ref := fs.String("ref", "", "agent reference")
		state := fs.String("state", "", "target state")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().TransitionAgent(ctx, *ref, *state, c.mutation())
		if err != nil {
			return err
		}
		return print(result)

	case "contain":
		fs := flag.NewFlagSet("agent contain", flag.ContinueOnError)
		ref := fs.String("ref", "", "agent reference")
		mode := fs.String("mode", "quarantine", "containment mode: quarantine or terminate")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().ContainAgent(ctx, *ref, *mode, c.mutation())
		if err != nil {
			return err
		}
		return print(result)

	default:
		return fmt.Errorf("unknown agent action %q", args[0])
	}
}

func runIdentity(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: nomivela identity <allocate|list|bind|lifecycle> [flags]")
	}
	c := newConfig()

	switch args[0] {
	case "allocate":
		fs := flag.NewFlagSet("identity allocate", flag.ContinueOnError)
		namespace := fs.String("namespace", "", "authority namespace")
		agentRef := fs.String("agent-ref", "", "agent reference")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().AllocateIdentity(ctx, *namespace, *agentRef, c.mutation())
		if err != nil {
			return err
		}
		return print(result)

	case "list":
		fs := flag.NewFlagSet("identity list", flag.ContinueOnError)
		namespace := fs.String("namespace", "", "authority namespace")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().ListIdentities(ctx, *namespace)
		if err != nil {
			return err
		}
		return print(result)

	case "bind":
		fs := flag.NewFlagSet("identity bind", flag.ContinueOnError)
		namespace := fs.String("namespace", "", "authority namespace")
		agentID := fs.String("agent-id", "", "agent id")
		rootRef := fs.String("root-ref", "", "authority root reference")
		rootType := fs.String("root-type", "organization", "authority root type")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().BindIdentity(ctx, *namespace, *agentID, *rootRef, *rootType, c.mutation())
		if err != nil {
			return err
		}
		return print(result)

	case "lifecycle":
		fs := flag.NewFlagSet("identity lifecycle", flag.ContinueOnError)
		namespace := fs.String("namespace", "", "authority namespace")
		agentID := fs.String("agent-id", "", "agent id")
		state := fs.String("state", "", "target state")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().TransitionIdentity(ctx, *namespace, *agentID, *state, c.mutation())
		if err != nil {
			return err
		}
		return print(result)

	default:
		return fmt.Errorf("unknown identity action %q", args[0])
	}
}

func runWorkload(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: nomivela workload <create|list|lifecycle> [flags]")
	}
	c := newConfig()

	switch args[0] {
	case "create":
		fs := flag.NewFlagSet("workload create", flag.ContinueOnError)
		namespace := fs.String("namespace", "", "authority namespace")
		platform := fs.String("platform", "", "platform")
		selector := fs.String("selector", "", "comma-separated k=v selector")
		trustDomain := fs.String("trust-domain", "", "trust domain")
		proofMethods := fs.String("proof-methods", "spiffe", "comma-separated proof methods")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		parsedSelector, err := parseSelector(*selector)
		if err != nil {
			return err
		}
		result, err := c.client().CreateWorkloadRegistration(ctx, nomivela.WorkloadRegistration{
			Namespace: *namespace, Platform: *platform, Selector: parsedSelector,
			TrustDomain: *trustDomain, AllowedProofMethods: splitList(*proofMethods),
		}, c.mutation())
		if err != nil {
			return err
		}
		return print(result)

	case "list":
		fs := flag.NewFlagSet("workload list", flag.ContinueOnError)
		namespace := fs.String("namespace", "", "authority namespace")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().ListWorkloadRegistrations(ctx, *namespace)
		if err != nil {
			return err
		}
		return print(result)

	case "lifecycle":
		fs := flag.NewFlagSet("workload lifecycle", flag.ContinueOnError)
		id := fs.String("id", "", "workload registration id")
		state := fs.String("state", "", "target state")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().TransitionWorkloadRegistration(ctx, *id, *state, c.mutation())
		if err != nil {
			return err
		}
		return print(result)

	default:
		return fmt.Errorf("unknown workload action %q", args[0])
	}
}

func runInstance(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: nomivela instance <commit|list|lifecycle> [flags]")
	}
	c := newConfig()

	switch args[0] {
	case "commit":
		fs := flag.NewFlagSet("instance commit", flag.ContinueOnError)
		namespace := fs.String("namespace", "", "authority namespace")
		agentID := fs.String("agent-id", "", "agent id")
		workloadRegistrationID := fs.String("workload-registration-id", "", "workload registration id")
		workloadID := fs.String("workload-id", "", "workload id")
		artifactDigest := fs.String("artifact-digest", "", "artifact digest")
		attestationRef := fs.String("attestation-ref", "", "verified attestation reference")
		lease := fs.String("lease-expires-at", "", "RFC 3339 lease expiry")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		leaseExpiry, err := time.Parse(time.RFC3339, *lease)
		if err != nil {
			return fmt.Errorf("--lease-expires-at must be RFC 3339: %w", err)
		}
		result, err := c.client().CommitInstance(ctx, *agentID, nomivela.InstanceCommit{
			Namespace: *namespace, WorkloadRegistrationID: *workloadRegistrationID, WorkloadID: *workloadID,
			ArtifactDigest: *artifactDigest, AttestationRef: *attestationRef, LeaseExpiresAt: leaseExpiry,
		}, c.mutation())
		if err != nil {
			return err
		}
		return print(result)

	case "list":
		fs := flag.NewFlagSet("instance list", flag.ContinueOnError)
		namespace := fs.String("namespace", "", "authority namespace")
		agentID := fs.String("agent-id", "", "agent id")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().ListInstances(ctx, *namespace, *agentID)
		if err != nil {
			return err
		}
		return print(result)

	case "lifecycle":
		fs := flag.NewFlagSet("instance lifecycle", flag.ContinueOnError)
		instanceID := fs.String("instance-id", "", "instance id")
		state := fs.String("state", "", "target state")
		c.globalFlags(fs)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := c.client().TransitionInstance(ctx, *instanceID, *state, c.mutation())
		if err != nil {
			return err
		}
		return print(result)

	default:
		return fmt.Errorf("unknown instance action %q", args[0])
	}
}

func runEvidence(ctx context.Context, args []string) error {
	c := newConfig()
	fs := flag.NewFlagSet("evidence", flag.ContinueOnError)
	c.globalFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := c.client().ListEvidence(ctx)
	if err != nil {
		return err
	}
	return print(result)
}

func runEvents(ctx context.Context, args []string) error {
	c := newConfig()
	fs := flag.NewFlagSet("events", flag.ContinueOnError)
	c.globalFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := c.client().ListEvents(ctx)
	if err != nil {
		return err
	}
	return print(result)
}

func print(value any) error {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func parseSelector(raw string) (map[string]string, error) {
	selector := map[string]string{}
	if strings.TrimSpace(raw) == "" {
		return selector, nil
	}
	for _, pair := range strings.Split(raw, ",") {
		key, value, ok := strings.Cut(pair, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("selector %q must be k=v", pair)
		}
		selector[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return selector, nil
}

func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
