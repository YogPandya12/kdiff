# kdelta Architecture

This document outlines the internal architecture and data flow of the `kdelta` CLI tool.

## System Architecture

The following diagram illustrates the core components of `kdelta` and how they interact based on different user commands.

```mermaid
graph TD
    %% User Inputs
    User([User CLI Execution])
    Files[/Local YAML Files/]
    KubeConfig[/Kubeconfig/]

    %% CLI Layer
    subgraph CMD [cmd/kdelta - CLI Layer]
        CommandRoute{Command Router}
        CmdCompare(compare)
        CmdValidate(validate)
        CmdLive(live)
        CmdDrift(drift)
        CmdCluster(cluster)
    end

    %% Core Packages
    subgraph PKG [pkg/ - Core Logic]
        Parser(pkg/parser<br/>YAML Parsing)
        Diff(pkg/diff<br/>Semantic Diffing)
        ValEngine(pkg/validation<br/>OpenAPI Validator)
        
        subgraph KUBE [pkg/kube - Kubernetes Integration]
            KubeClient(Client Init)
            ResourceFetch(Resource Fetcher)
            Normalize(Normalizer)
        end
        
        History(pkg/history<br/>Drift Logger)
    end

    %% External
    Cluster1[(Kubernetes Cluster 1)]
    Cluster2[(Kubernetes Cluster 2)]
    OpenAPISchema[(K8s OpenAPI Schemas)]
    DriftLog[/drift.log/]

    %% Connections
    User --> CommandRoute
    CommandRoute --> CmdCompare
    CommandRoute --> CmdValidate
    CommandRoute --> CmdLive
    CommandRoute --> CmdDrift
    CommandRoute --> CmdCluster

    %% Compare Flow
    CmdCompare -->|Reads| Files
    CmdCompare --> Parser
    Parser --> Diff

    %% Validate Flow
    CmdValidate -->|Reads| Files
    CmdValidate --> Parser
    CmdValidate --> ValEngine
    ValEngine -->|Fetches/Uses| OpenAPISchema

    %% Live / Drift Flow
    CmdLive & CmdDrift -->|Reads| Files
    CmdLive & CmdDrift --> Parser
    CmdLive & CmdDrift --> KubeClient
    KubeClient -->|Uses| KubeConfig
    KubeClient --> ResourceFetch
    ResourceFetch -->|Queries| Cluster1
    ResourceFetch --> Normalize
    Normalize --> Diff
    CmdDrift --> History
    History -->|Writes| DriftLog

    %% Cluster Flow
    CmdCluster --> KubeClient
    KubeClient -->|Fetches from Context 1| Cluster1
    KubeClient -->|Fetches from Context 2| Cluster2
    CmdCluster --> Normalize
    Normalize --> Diff

    %% Styling
    classDef cmd fill:#f9f,stroke:#333,stroke-width:2px;
    classDef pkg fill:#bbf,stroke:#333,stroke-width:2px;
    classDef ext fill:#dfd,stroke:#333,stroke-width:2px;

    class CmdCompare,CmdValidate,CmdLive,CmdDrift,CmdCluster cmd;
    class Parser,Diff,ValEngine,KubeClient,ResourceFetch,Normalize,History pkg;
    class Cluster1,Cluster2,OpenAPISchema ext;
```

## Component Breakdown

1.  **`cmd/kdelta`**: Uses Cobra to handle command-line arguments, routing, and formatting the final output (JSON, Table, or colored text).
2.  **`pkg/parser`**: Responsible for reading YAML files and converting them into canonical, unstructured objects that the diffing engine can understand.
3.  **`pkg/diff`**: The core line-by-line comparison engine. It compares canonicalized YAML representations to identify semantic additions, deletions, and modifications.
4.  **`pkg/validation`**: Interfaces with the `go-openapi` libraries to validate parsed YAML files against official Kubernetes OpenAPI definitions.
5.  **`pkg/kube`**: 
    *   Initializes the `client-go` dynamic client.
    *   Fetches live resources from the cluster.
    *   **Normalizer**: Strips out runtime-specific fields (like `status`, `uid`, `creationTimestamp`) before comparison to ensure accurate diffing.
6.  **`pkg/history`**: Handles appending drift events in JSON format to the local audit log (`drift.log`).
