import time
import json
import statistics
import urllib.request
import urllib.error

def benchmark_endpoint(url, payload, n_requests=300):
    latencies = []
    errors = 0
    degraded_count = 0
    
    # Warmup
    for _ in range(10):
        try:
            req = urllib.request.Request(
                url,
                data=json.dumps(payload).encode('utf-8'),
                headers={'Content-Type': 'application/json'}
            )
            with urllib.request.urlopen(req, timeout=2.0) as resp:
                _ = resp.read()
        except Exception:
            pass

    for i in range(n_requests):
        t0 = time.perf_counter()
        try:
            req = urllib.request.Request(
                url,
                data=json.dumps(payload).encode('utf-8'),
                headers={'Content-Type': 'application/json'}
            )
            with urllib.request.urlopen(req, timeout=2.0) as resp:
                body = resp.read()
                data = json.loads(body.decode('utf-8'))
                if data.get("is_degraded") is True:
                    degraded_count += 1
            latencies.append((time.perf_counter() - t0) * 1000.0) # ms
        except Exception as e:
            errors += 1
            
    if not latencies:
        return {"error": "All requests failed", "errors": errors}
        
    latencies.sort()
    p50 = statistics.median(latencies)
    p95 = latencies[int(len(latencies) * 0.95)]
    p99 = latencies[int(len(latencies) * 0.99)]
    mean = statistics.mean(latencies)
    min_val = min(latencies)
    max_val = max(latencies)
    
    return {
        "requests": n_requests,
        "successful": len(latencies),
        "errors": errors,
        "degraded_count": degraded_count,
        "degraded_rate": round(degraded_count / len(latencies), 4) if latencies else 0,
        "p50_ms": round(p50, 2),
        "p95_ms": round(p95, 2),
        "p99_ms": round(p99, 2),
        "mean_ms": round(mean, 2),
        "min_ms": round(min_val, 2),
        "max_ms": round(max_val, 2)
    }

if __name__ == "__main__":
    print("==================================================")
    print("MEASURING ACTUAL LIVE PERFORMANCE BASELINE")
    print("==================================================")
    
    # 1. Benchmark Go API Synchronous Decision Pipeline
    api_url = "http://localhost:8080/v1/risk-evaluations"
    api_payload = {
        "amount": 15000,
        "currency": "INR",
        "payment_method": {
            "type": "card",
            "token": "tok_bench_4242"
        },
        "device_fingerprint": "fp_bench_client_test",
        "ip_address": "198.51.100.77"
    }
    
    print("Benchmarking Go Decision API (300 requests)...")
    api_results = benchmark_endpoint(api_url, api_payload, n_requests=300)
    print(f"API Latency: p50={api_results['p50_ms']}ms, p95={api_results['p95_ms']}ms, p99={api_results['p99_ms']}ms (Errors: {api_results['errors']}, Degraded: {api_results['degraded_count']})")
    
    # 2. Benchmark ML Sidecar directly
    ml_url = "http://localhost:8000/predict"
    ml_payload = {
        "amount": 15000.0,
        "ip_velocity_1h": 2.0,
        "token_velocity_24h": 3.0,
        "is_new_device": 0,
        "hour_of_day": 14
    }
    
    print("Benchmarking Python ONNX ML Sidecar (300 requests)...")
    ml_results = benchmark_endpoint(ml_url, ml_payload, n_requests=300)
    print(f"ML Latency: p50={ml_results['p50_ms']}ms, p95={ml_results['p95_ms']}ms, p99={ml_results['p99_ms']}ms (Errors: {ml_results['errors']})")
    
    combined = {
        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "sample_size": 300,
        "api_evaluation": api_results,
        "ml_sidecar": ml_results
    }
    
    with open("ml-service/evaluation/live_performance_benchmark.json", "w") as f:
        json.dump(combined, f, indent=2)
    print("Saved live benchmark results to ml-service/evaluation/live_performance_benchmark.json")
