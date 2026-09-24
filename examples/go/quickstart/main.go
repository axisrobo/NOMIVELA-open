// Command quickstart runs the full NOMIVELA registration and enrollment path
// against a running NOMIVELA Core.
//
//	go run ./examples/go/quickstart -api http://localhost:8080
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	nomivela "github.com/axisrobo/nomivela-open/sdk/go"
)

func main() {
	api := flag.String("api", "http://localhost:8080", "API base URL")
	namespace := flag.String("namespace", "https://auth.example.com", "authority namespace")
	flag.Parse()

	ctx := context.Background()
	client := nomivela.New(*api, nomivela.WithActor("example:quickstart"))
	mutation := nomivela.Mutation{Reason: "quickstart"}

	ns, err := client.CreateNamespace(ctx, *namespace, "root:org-a", mutation)
	if err != nil {
		log.Fatalf("create namespace: %v", err)
	}
	fmt.Printf("namespace %s (%s, epoch %d)\n", ns.Namespace, ns.Status, ns.NamespaceEpoch)

	agent, err := client.CreateAgent(ctx, nomivela.Agent{
		AgentRef: "agent_order_processor", Name: "Order Processor", Purpose: "Process approved order events",
		SponsorRef: "org:platform", OwnerRef: "user:owner", RiskClass: "medium",
	}, mutation)
	if err != nil {
		log.Fatalf("create agent: %v", err)
	}
	for _, state := range []string{"registered", "approved", "active"} {
		if agent, err = client.TransitionAgent(ctx, agent.AgentRef, state, mutation); err != nil {
			log.Fatalf("agent %s: %v", state, err)
		}
	}
	fmt.Printf("agent %s (%s, epoch %d)\n", agent.AgentRef, agent.State, agent.AgentEpoch)

	identity, err := client.AllocateIdentity(ctx, *namespace, agent.AgentRef, mutation)
	if err != nil {
		log.Fatalf("allocate identity: %v", err)
	}
	if identity, err = client.BindIdentity(ctx, *namespace, identity.AgentID, "root:org-a", "organization", mutation); err != nil {
		log.Fatalf("bind identity: %v", err)
	}
	if identity, err = client.TransitionIdentity(ctx, *namespace, identity.AgentID, "active", mutation); err != nil {
		log.Fatalf("activate identity: %v", err)
	}
	fmt.Printf("identity %s (%s, epoch %d)\n", identity.AgentID, identity.State, identity.IdentityEpoch)

	workload, err := client.CreateWorkloadRegistration(ctx, nomivela.WorkloadRegistration{
		Namespace: *namespace, Platform: "kubernetes", Selector: map[string]string{"app": "orders"},
		TrustDomain: "cluster.local", AllowedProofMethods: []string{"spiffe"},
	}, mutation)
	if err != nil {
		log.Fatalf("create workload: %v", err)
	}
	fmt.Printf("workload %s (%s)\n", workload.WorkloadRegistrationID, workload.Status)

	instance, err := client.CommitInstance(ctx, identity.AgentID, nomivela.InstanceCommit{
		Namespace: *namespace, WorkloadRegistrationID: workload.WorkloadRegistrationID,
		WorkloadID: "workload-1", ArtifactDigest: "sha256:abc", AttestationRef: "attestation:1",
		LeaseExpiresAt: time.Now().Add(time.Hour),
	}, mutation)
	if err != nil {
		log.Fatalf("commit instance: %v", err)
	}
	fmt.Printf("instance %s (%s)\n", instance.InstanceID, instance.State)

	evidence, err := client.ListEvidence(ctx)
	if err != nil {
		log.Fatalf("list evidence: %v", err)
	}
	fmt.Printf("evidence records: %d (last: %s -> %s by %s)\n",
		len(evidence), evidence[len(evidence)-1].PreviousState, evidence[len(evidence)-1].NewState, evidence[len(evidence)-1].Actor)
}
