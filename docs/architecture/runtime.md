### 1. Difference Between Configuration and Runtime

* **Configuration** refers to settings that determine how an application should behave. These are typically provided before or during startup, such as environment variables, configuration files, or command-line arguments.
    
* **Runtime** refers to what happens while the application is actively running. It includes the application's current state, resource usage, and execution behavior.

**In short:** Configuration tells the application *how to run*; runtime describes *what is happening while it runs*.

---

### 2. Why Uptime Matters

**Uptime** is the amount of time an application has been running continuously since it last started.

It is useful because it helps:

* Detect unexpected restarts or crashes.
* Verify application stability.
* Troubleshoot issues after deployments.
* Monitor service reliability.
* Determine whether configuration changes required a restart.

For example:

* Uptime: **15 days** → application has been stable.
* Uptime: **2 minutes** → application likely just restarted.

---

### 3. Purpose of `/status`

A `/status` endpoint (sometimes called `/health` or `/info`) provides basic information about the application's current state.

Typical information includes:

* Service is running
* Application version
* Environment (development, staging, production)
* Uptime
* Build information
* Current timestamp
* Health of dependencies (optional)



This endpoint is commonly used by:

* Load balancers
* Monitoring systems
* Kubernetes health checks
* Developers diagnosing deployments

---

### 4. Future Build Metadata Injection

Build metadata injection is the practice of embedding information about a specific build into the application during the build or deployment process instead of hardcoding it.

