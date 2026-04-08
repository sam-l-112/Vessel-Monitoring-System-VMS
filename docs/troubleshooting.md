# Troubleshooting Guide

This guide provides solutions to common issues encountered during development, testing, and deployment of the Vessel Monitoring System (VMS).

## Docker-Related Issues

### Port Already in Use
**Error:** `bind: address already in use`
**Symptoms:** Docker containers fail to start with port binding errors.
**Solutions:**
1. Check which process is using the port:
   ```bash
   sudo lsof -i :PORT_NUMBER
   ```
2. Stop the conflicting service or change the port in `docker-compose.yml`
3. Restart Docker services:
   ```bash
   docker-compose down
   docker-compose up -d
   ```

### Image Build Failures
**Error:** `Image not found` or build errors during `docker-compose build`
**Solutions:**
1. Clear Docker cache:
   ```bash
   docker system prune -a
   ```
2. Rebuild images:
   ```bash
   docker-compose build --no-cache
   ```
3. Check Dockerfile syntax and dependencies

### Container Startup Issues
**Symptoms:** Containers exit immediately after starting
**Solutions:**
1. Check container logs:
   ```bash
   docker-compose logs SERVICE_NAME
   ```
2. Verify environment variables in `docker-compose.yml`
3. Ensure required volumes and networks are properly configured

## Backend (Golang) Issues

### Database Connection Refused
**Error:** `dial tcp [::1]:3306: connect: connection refused`
**Solutions:**
1. Ensure MariaDB container is running:
   ```bash
   docker-compose ps
   ```
2. Check database credentials in configuration files
3. Verify database initialization scripts have run
4. Test connection manually:
   ```bash
   docker-compose exec mariadb mysql -u USER -p DATABASE
   ```

### API Endpoint Errors
**Symptoms:** HTTP 500 errors or unexpected responses
**Solutions:**
1. Check application logs:
   ```bash
   docker-compose logs backend
   ```
2. Verify API routes and handlers in Golang code
3. Test endpoints with curl or Postman
4. Check for nil pointer dereferences or unhandled errors

### OpenClaw CLI Integration Issues
**Symptoms:** External data not being retrieved
**Solutions:**
1. Verify OpenClaw CLI installation and configuration
2. Check API keys and authentication tokens
3. Test CLI commands manually outside the application
4. Review error handling for external API calls

## Frontend (Vue.js) Issues

### Module Not Found Errors
**Error:** `Cannot resolve module 'X'`
**Solutions:**
1. Install missing dependencies:
   ```bash
   npm install
   ```
2. Clear node_modules and reinstall:
   ```bash
   rm -rf node_modules package-lock.json
   npm install
   ```
3. Check import paths in Vue components

### Build Failures
**Symptoms:** `npm run build` fails
**Solutions:**
1. Check for TypeScript or ESLint errors
2. Verify Vue.js version compatibility
3. Update dependencies to compatible versions
4. Clear build cache:
   ```bash
   npm run clean
   ```

### Runtime Errors
**Symptoms:** JavaScript errors in browser console
**Solutions:**
1. Check browser developer tools for error messages
2. Verify API endpoints are accessible
3. Test with different browsers
4. Check for CORS issues if making cross-origin requests

## Database (MariaDB) Issues

### Connection Timeouts
**Symptoms:** Slow queries or connection timeouts
**Solutions:**
1. Optimize database queries
2. Check database indexes
3. Increase connection pool size in configuration
4. Monitor database performance with `SHOW PROCESSLIST`

### Data Corruption or Loss
**Symptoms:** Inconsistent or missing data
**Solutions:**
1. Restore from backup if available
2. Check database logs for errors
3. Run database integrity checks
4. Implement proper transaction handling in application code

## Python Algorithm Issues

### Import Errors
**Error:** `ModuleNotFoundError`
**Solutions:**
1. Install required Python packages:
   ```bash
   pip install -r requirements.txt
   ```
2. Check Python path and virtual environment
3. Verify package versions are compatible

### Algorithm Performance Issues
**Symptoms:** Slow data processing or high memory usage
**Solutions:**
1. Profile code with `cProfile` or similar tools
2. Optimize algorithms for better time/space complexity
3. Consider using NumPy/Pandas for data processing
4. Implement caching where appropriate

## General Debugging Steps

1. **Check Logs**
   ```bash
   # View all service logs
   docker-compose logs

   # Follow logs in real-time
   docker-compose logs -f SERVICE_NAME
   ```

2. **Verify Configuration**
   - Check environment variables
   - Validate configuration files
   - Ensure correct file permissions

3. **Test Individual Components**
   - Run unit tests: `go test ./...` (backend)
   - Test API endpoints manually
   - Verify database connections

4. **System Resource Check**
   ```bash
   # Check disk space
   df -h

   # Check memory usage
   free -h

   # Check Docker resources
   docker system df
   ```

5. **Restart Services**
   ```bash
   docker-compose restart
   ```

6. **Clean Environment**
   ```bash
   # Stop and remove containers
   docker-compose down -v

   # Remove unused resources
   docker system prune -a
   ```

## Getting Help

If issues persist:
1. Check the project's GitHub issues for similar problems
2. Provide detailed error logs and system information when seeking help
3. Include steps to reproduce the issue
4. Specify your environment (OS, Docker version, Go/Python versions)