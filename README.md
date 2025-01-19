## **12-Factor App Implementation**

This application follows the **12-Factor App methodology**, ensuring scalability, maintainability, and portability.  

## **1. Codebase**: One Codebase Tracked in Version Control  
- All code is contained in a **single Git repository**.  
- Managed using **Git** for version control.  

## **2. Dependencies**: Explicitly Declare and Isolate Dependencies  
- Uses **Go modules (`go.mod`)** to manage dependencies.  
- Dependencies are **explicitly versioned** and **isolated**.  

## **3. Config**: Store Configuration in the Environment  
- Configuration is managed via **environment variables**.  
- Supports `.env` files with **godotenv** for local development.  
- Examples include `DATABASE_URL`, `PORT`, `APP_ENV`.  

## **4. Backing Services**: Treat Backing Services as Attached Resources  
- Uses **PostgreSQL**, configured via environment variables.  
- Easily switch between **development, staging, and production databases**.  

## **5. Build, Release, Run**: Strictly Separate Build and Run Stages  
- **Dockerfile** implements a **multi-stage build** for smaller, efficient images.  
- Clear separation of **build**, **release**, and **run** phases.  

## **6. Processes**: Execute the App as One or More Stateless Processes  
- The application is stateless; **no local state** is stored.  
- Persistent data is stored in the **database** instead of the filesystem.  

## **7. Port Binding**: Export Services via Port Binding  
- The application binds to a **configurable port** (`PORT` environment variable).  
- Ready for **containerized deployment** in Docker and Kubernetes.  

## **8. Concurrency**: Scale Out via the Process Model  
- Uses **Gin framework** to efficiently handle concurrent requests.  
- Stateless design enables **horizontal scaling**.  

## **9. Disposability**: Maximize Robustness with Fast Startup and Graceful Shutdown  
- Supports **graceful shutdown** using **Gin's built-in capabilities**.  
- Ensures fast startup and avoids long initialization times.  

## **10. Dev/Prod Parity**: Keep Development, Staging, and Production as Similar as Possible  
- Uses **Docker** and **docker-compose** for a consistent development environment.  
- Configuration remains **identical across environments**.  

## **11. Logs**: Treat Logs as Event Streams  
- Uses structured logging with **Gin's built-in logger**.  
- Logs are streamed to **stdout**, allowing easy collection by infrastructure tools.  



### Kubernetes deploy

```sh
helm upgrade --install mailping ./mailping-deploy --create-namespace --namespace mailping
```


### ToDo

1. [Instrumenting Prometheus](https://prometheus.io/docs/guides/go-application/#instrumenting-a-go-application-for-prometheus)
2. Expose tracking stats as Prometheus metrics