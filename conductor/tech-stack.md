# Technology Stack

## Backend
- **Go (v1.25.0):** Chosen for its performance, concurrency support, and strong tooling for infrastructure-related tasks.
- **Gin:** A high-performance web framework used for its simplicity and robustness in building RESTful APIs and UI routing.
- **PostgreSQL:** The primary relational database for robust, ACID-compliant storage of system configuration and deployment history.

## Frontend
- **Tailwind CSS (v3.4.1):** Utilized for a modern, responsive, and highly-customizable user interface through a utility-first CSS approach.
- **HTMX:** Leveraged to create interactive UI components with minimal JavaScript by using HTML attributes to trigger AJAX requests. (Inferred from project goals and plans)

## Infrastructure & Orchestration
- **Terraform:** The core tool for defining and provisioning the Proxmox virtual infrastructure through "Infrastructure as Code" (IaC).
- **Proxmox Virtual Environment:** The target hypervisor for managing the compute layer (Virtual Machines).
- **Docker:** Used to containerize and deploy microservices onto the dynamically provisioned Virtual Machines.

## Tooling
- **git:** The primary version control system for tracking codebase changes and providing a commit-based history for deployments.
- **Trivy:** Integrated for automated vulnerability scanning of Docker images (Inferred from plans).
- **VictoriaMetrics:** Employed for efficient, long-term storage and real-time streaming of performance metrics (Inferred from plans).
