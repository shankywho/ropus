import json

def make_rect(id, x, y, w, h, bg="#1e1e2e", stroke="#89b4fa", label="", font_size=14, stroke_width=2, roundness={"type": 3}, text_align="center", text_color="#cdd6f4", font_family=3):
    elements = []
    rect = {
        "id": id,
        "type": "rectangle",
        "x": x,
        "y": y,
        "width": w,
        "height": h,
        "angle": 0,
        "strokeColor": stroke,
        "backgroundColor": bg,
        "fillStyle": "solid",
        "strokeWidth": stroke_width,
        "strokeStyle": "solid",
        "roughness": 0,
        "opacity": 100,
        "groupIds": [],
        "frameId": None,
        "roundness": roundness,
        "seed": 1000 + len(id) * 31,
        "version": 1,
        "versionNonce": 1,
        "isDeleted": False,
        "boundElements": None,
        "updated": 1,
        "link": None,
        "locked": False
    }
    elements.append(rect)
    
    if label:
        text_id = f"text_{id}"
        lines = label.split("\n")
        line_height = font_size * 1.35
        text_h = len(lines) * line_height
        text_y = y + (h - text_h) / 2
        text_x = x + 12
        
        text_elem = {
            "id": text_id,
            "type": "text",
            "x": text_x,
            "y": text_y,
            "width": w - 24,
            "height": text_h,
            "angle": 0,
            "strokeColor": text_color,
            "backgroundColor": "transparent",
            "fillStyle": "solid",
            "strokeWidth": 1,
            "strokeStyle": "solid",
            "roughness": 0,
            "opacity": 100,
            "groupIds": [],
            "frameId": None,
            "roundness": None,
            "seed": 2000 + len(id) * 37,
            "version": 1,
            "versionNonce": 1,
            "isDeleted": False,
            "boundElements": None,
            "updated": 1,
            "link": None,
            "locked": False,
            "text": label,
            "fontSize": font_size,
            "fontFamily": font_family, # 3: Monospace
            "textAlign": text_align,
            "verticalAlign": "middle",
            "containerId": id,
            "originalText": label,
            "lineHeight": 1.35
        }
        elements.append(text_elem)
        rect["boundElements"] = [{"id": text_id, "type": "text"}]
        
    return elements

def make_text_standalone(id, x, y, text, font_size=15, color="#cba6f7", font_family=3, align="left"):
    lines = text.split("\n")
    line_height = font_size * 1.35
    text_h = len(lines) * line_height
    return [{
        "id": id,
        "type": "text",
        "x": x,
        "y": y,
        "width": 1400,
        "height": text_h,
        "angle": 0,
        "strokeColor": color,
        "backgroundColor": "transparent",
        "fillStyle": "solid",
        "strokeWidth": 1,
        "strokeStyle": "solid",
        "roughness": 0,
        "opacity": 100,
        "groupIds": [],
        "frameId": None,
        "roundness": None,
        "seed": 5000 + len(id),
        "version": 1,
        "versionNonce": 1,
        "isDeleted": False,
        "boundElements": None,
        "updated": 1,
        "link": None,
        "locked": False,
        "text": text,
        "fontSize": font_size,
        "fontFamily": font_family,
        "textAlign": align,
        "verticalAlign": "top",
        "containerId": None,
        "originalText": text,
        "lineHeight": 1.35
    }]

def make_arrow(id, start_x, start_y, end_x, end_y, stroke="#89b4fa"):
    dx = end_x - start_x
    dy = end_y - start_y
    return [{
        "id": id,
        "type": "arrow",
        "x": start_x,
        "y": start_y,
        "width": max(abs(dx), 1),
        "height": max(abs(dy), 1),
        "angle": 0,
        "strokeColor": stroke,
        "backgroundColor": "transparent",
        "fillStyle": "solid",
        "strokeWidth": 2,
        "strokeStyle": "solid",
        "roughness": 0,
        "opacity": 100,
        "groupIds": [],
        "frameId": None,
        "roundness": {"type": 2},
        "seed": 3000 + len(id),
        "version": 1,
        "versionNonce": 1,
        "isDeleted": False,
        "boundElements": None,
        "updated": 1,
        "link": None,
        "locked": False,
        "points": [
            [0, 0],
            [dx, dy]
        ],
        "lastCommittedPoint": None,
        "startBinding": None,
        "endBinding": None,
        "startArrowhead": None,
        "endArrowhead": "arrow"
    }]

def build_excalidraw():
    all_elements = []
    
    # ----------------------------------------------------
    # Header Banner (Y: 30, H: 80)
    # ----------------------------------------------------
    all_elements.extend(make_rect(
        "title_box", 40, 30, 1600, 80,
        bg="#181825", stroke="#cba6f7",
        label="ROPUS — Autonomous AI Risk Manager & Financial Crime Intelligence Platform\nEnd-to-End System Architecture Specification",
        font_size=18, stroke_width=2, text_color="#cba6f7"
    ))

    # ----------------------------------------------------
    # LAYER 1: CLIENT INGESTION & INTERACTION CHANNELS
    # Y: 140, H: 180 (Header Y: 155, Cards Y: 185)
    # ----------------------------------------------------
    all_elements.extend(make_rect("grp_l1", 40, 140, 1600, 180, bg="#11111b", stroke="#313244", stroke_width=1.5))
    all_elements.extend(make_text_standalone("lbl_l1", 60, 155, "LAYER 1: CLIENT INGESTION & INTERACTION CHANNELS", font_size=13, color="#89b4fa"))
    
    all_elements.extend(make_rect("c1", 65, 185, 360, 115, bg="#1e1e2e", stroke="#89b4fa", 
        label="Customer Inbound REST API\n• Endpoint: POST /v1/risk/evaluate\n• Payload: Tx, Device, Geo, IP Telemetry\n• Synchronous Decision Response", font_size=12))
    
    all_elements.extend(make_rect("c2", 460, 185, 360, 115, bg="#1e1e2e", stroke="#89b4fa",
        label="Analyst Web Portal (Next.js)\n• 18 Production Console Routes\n• Graph Explorer, Case Queue, Governance\n• Live WebSocket Event Ticker", font_size=12))
    
    all_elements.extend(make_rect("c3", 855, 185, 360, 115, bg="#1e1e2e", stroke="#89b4fa",
        label="Client SDKs (Python & Node.js)\n• Drop-in integration in < 15 minutes\n• Automated HMAC request signing\n• In-memory client connection pooling", font_size=12))
    
    all_elements.extend(make_rect("c4", 1250, 185, 360, 115, bg="#1e1e2e", stroke="#89b4fa",
        label="Customer Webhook Configuration\n• Endpoint & Secret Key Management\n• Subscriptions (risk.decision, case.p0)\n• Custom Event Filter Rules", font_size=12))

    # Arrows L1 -> L2 (Downwards)
    all_elements.extend(make_arrow("arr_1_1", 245, 300, 245, 360, stroke="#89b4fa"))
    all_elements.extend(make_arrow("arr_1_2", 640, 300, 640, 360, stroke="#89b4fa"))
    all_elements.extend(make_arrow("arr_1_3", 1035, 300, 1035, 360, stroke="#89b4fa"))
    all_elements.extend(make_arrow("arr_1_4", 1430, 300, 1430, 360, stroke="#89b4fa"))

    # ----------------------------------------------------
    # LAYER 2: ZERO-TRUST SECURITY, AUTH & MULTI-TENANT GATEWAY
    # Y: 360, H: 180 (Header Y: 375, Cards Y: 405)
    # ----------------------------------------------------
    all_elements.extend(make_rect("grp_l2", 40, 360, 1600, 180, bg="#11111b", stroke="#313244", stroke_width=1.5))
    all_elements.extend(make_text_standalone("lbl_l2", 60, 375, "LAYER 2: ZERO-TRUST SECURITY, AUTHENTICATION & MULTI-TENANT GATEWAY", font_size=13, color="#f38ba8"))

    all_elements.extend(make_rect("g1", 65, 405, 360, 115, bg="#1e1e2e", stroke="#f38ba8",
        label="API Key Authentication\n• SHA-256 One-Way Key Hash Vault\n• Prefixes: rop_live_ / rop_test_\n• Zero Plaintext Storage in DB", font_size=12))
    
    all_elements.extend(make_rect("g2", 460, 405, 360, 115, bg="#1e1e2e", stroke="#f38ba8",
        label="Request Validation & WAF\n• SQLi / XSS Parameter Sanitization\n• Strict Schema Typing & Normalization\n• Replay Attack Timestamp Guard", font_size=12))
    
    all_elements.extend(make_rect("g3", 855, 405, 360, 115, bg="#1e1e2e", stroke="#f9e2af",
        label="Distributed Rate Limiter\n• Token Bucket with 2x Burst Allowance\n• Starter: 100 | Growth: 500 | Ent: 5,000 RPS\n• Strict Tenant Quota Enforcement", font_size=12))
    
    all_elements.extend(make_rect("g4", 1250, 405, 360, 115, bg="#1e1e2e", stroke="#cba6f7",
        label="Tenant Isolation & RBAC\n• 4-Tier Roles (Owner, Admin, Analyst, Viewer)\n• Organization-Level Boundary Isolation\n• Atomic Usage Metering for Billing", font_size=12))

    # Arrows L2 -> L3 (Downwards)
    all_elements.extend(make_arrow("arr_2_1", 245, 520, 210, 580, stroke="#89b4fa"))
    all_elements.extend(make_arrow("arr_2_2", 640, 520, 520, 580, stroke="#89b4fa"))
    all_elements.extend(make_arrow("arr_2_3", 1035, 520, 1140, 580, stroke="#89b4fa"))
    all_elements.extend(make_arrow("arr_2_4", 1430, 520, 1450, 580, stroke="#89b4fa"))

    # ----------------------------------------------------
    # LAYER 3: REAL-TIME RISK INTELLIGENCE (<10ms TARGET)
    # Y: 580, H: 200 (Header Y: 595, Cards Y: 625)
    # ----------------------------------------------------
    all_elements.extend(make_rect("grp_l3", 40, 580, 1600, 200, bg="#11111b", stroke="#313244", stroke_width=1.5))
    all_elements.extend(make_text_standalone("lbl_l3", 60, 595, "LAYER 3: REAL-TIME RISK INTELLIGENCE & FEATURE ENGINES (<10ms TARGET)", font_size=13, color="#89dceb"))

    all_elements.extend(make_rect("e1", 65, 625, 285, 135, bg="#1e1e2e", stroke="#89dceb",
        label="Rules Engine\n• Declarative Policy Rules\n• Hard Velocity Ceilings\n• Geofencing & Sanctions Check\n• Instant Step-Up Triggering", font_size=11))
    
    all_elements.extend(make_rect("e2", 370, 625, 290, 135, bg="#1e1e2e", stroke="#a6e3a1",
        label="ML Inference Engine\n• Gradient Boosted Trees (XGBoost)\n• Continuous Probabilities (0.0-1.0)\n• Multi-Variate Anomaly Scoring\n• In-Memory Fast Matrix Inference", font_size=11))
    
    all_elements.extend(make_rect("e3", 680, 625, 290, 135, bg="#1e1e2e", stroke="#f38ba8",
        label="Fraud Knowledge Graph 3.0\n• In-Memory 3-Hop Traversal\n• Shared Device / Identity Links\n• Syndicate Cluster Discovery\n• Entity Degree & PageRank", font_size=11))
    
    all_elements.extend(make_rect("e4", 990, 625, 290, 135, bg="#1e1e2e", stroke="#fab387",
        label="Threat Intelligence\n• Bulletproof VPN/Proxy Subnets\n• Impossible Travel (>5k km jumps)\n• Device Canvas & Telemetry\n• Autonomous Subnet Blacklisting", font_size=11))
    
    all_elements.extend(make_rect("e5", 1300, 625, 310, 135, bg="#1e1e2e", stroke="#cba6f7",
        label="Behavioral & Device Signals\n• Historical Velocity Baseline\n• Device Canvas Fingerprint Drift\n• Expenditure Variance Profiling\n• Session Telemetry Tracking", font_size=11))

    # Arrows L3 -> L4 (Downwards)
    all_elements.extend(make_arrow("arr_3_1", 207, 760, 245, 820, stroke="#cdd6f4"))
    all_elements.extend(make_arrow("arr_3_2", 515, 760, 640, 820, stroke="#cdd6f4"))
    all_elements.extend(make_arrow("arr_3_3", 825, 760, 825, 820, stroke="#cdd6f4"))
    all_elements.extend(make_arrow("arr_3_4", 1135, 760, 1035, 820, stroke="#cdd6f4"))
    all_elements.extend(make_arrow("arr_3_5", 1455, 760, 1430, 820, stroke="#cdd6f4"))

    # ----------------------------------------------------
    # LAYER 4: COMPOSITE RISK DECISION & POLICY THRESHOLDING
    # Y: 820, H: 180 (Header Y: 835, Cards Y: 865)
    # ----------------------------------------------------
    all_elements.extend(make_rect("grp_l4", 40, 820, 1600, 180, bg="#11111b", stroke="#313244", stroke_width=1.5))
    all_elements.extend(make_text_standalone("lbl_l4", 60, 835, "LAYER 4: COMPOSITE RISK DECISION & ATTRIBUTION ENGINE", font_size=13, color="#a6e3a1"))

    all_elements.extend(make_rect("d1", 65, 865, 360, 115, bg="#1e1e2e", stroke="#a6e3a1",
        label="APPROVE Verdict\n• Score < 0.30 | High Confidence (>0.95)\n• Frictionless Immediate Settlement\n• Emits Background Event to Telemetry Bus", font_size=12))
    
    all_elements.extend(make_rect("d2", 460, 865, 360, 115, bg="#1e1e2e", stroke="#f9e2af",
        label="REVIEW / CHALLENGE Verdict\n• Score 0.30 - 0.79 | Medium Risk\n• Adaptive WebAuthn / Step-Up MFA\n• Queued for Analyst Review", font_size=12))
    
    all_elements.extend(make_rect("d3", 855, 865, 360, 115, bg="#1e1e2e", stroke="#f38ba8",
        label="BLOCK Verdict\n• Score >= 0.80 | Malicious Activity\n• Automated Settlement Halt & Session Freeze\n• Immediate P0 Case Creation Trigger", font_size=12))
    
    all_elements.extend(make_rect("d4", 1250, 865, 360, 115, bg="#1e1e2e", stroke="#89b4fa",
        label="Transparent Factor Attribution\n• Exact Additive Contributions (Sum == Score)\n• Measured Integration Latency: 1.42ms avg\n• Load-Test P99: 6.80ms | JSON Payload Return", font_size=12))

    # Arrows L4 -> L5 (Downwards)
    all_elements.extend(make_arrow("arr_4_1", 245, 980, 245, 1040, stroke="#a6e3a1"))
    all_elements.extend(make_arrow("arr_4_2", 640, 980, 640, 1040, stroke="#f9e2af"))
    all_elements.extend(make_arrow("arr_4_3", 1035, 980, 1035, 1040, stroke="#f38ba8"))
    all_elements.extend(make_arrow("arr_4_4", 1430, 980, 1430, 1040, stroke="#cba6f7"))

    # ----------------------------------------------------
    # LAYER 5: INVESTIGATION & GOVERNANCE WORKFLOW
    # Y: 1040, H: 180 (Header Y: 1055, Cards Y: 1085)
    # ----------------------------------------------------
    all_elements.extend(make_rect("grp_l5", 40, 1040, 1600, 180, bg="#11111b", stroke="#313244", stroke_width=1.5))
    all_elements.extend(make_text_standalone("lbl_l5", 60, 1055, "LAYER 5: AUTONOMOUS AI INVESTIGATION & GOVERNANCE WORKFLOW", font_size=13, color="#cba6f7"))

    all_elements.extend(make_rect("i1", 65, 1085, 360, 115, bg="#1e1e2e", stroke="#cba6f7",
        label="AI Investigator Agent\n• Multi-LLM Gateway (Claude 3.7 / GPT-4o)\n• Agent Council Multi-Persona Consensus\n• Token Cost Accounting & Smart Routing", font_size=12))
    
    all_elements.extend(make_rect("i2", 460, 1085, 360, 115, bg="#1e1e2e", stroke="#cba6f7",
        label="Structured Evidentiary Dossier\n• 1. Observed Facts (Verifiable Logs)\n• 2. Inferred Patterns (Transnational Rings)\n• 3. Recommended Actions (SAR Package)", font_size=12))
    
    all_elements.extend(make_rect("i3", 855, 1085, 360, 115, bg="#1e1e2e", stroke="#89dceb",
        label="Case Management & Audit Ledger\n• Priority P0 Case #CASE-88419\n• Immutable SHA-256 Hash-Chained Audit Trail\n• Attached Graph Artifacts & Timeline", font_size=12))
    
    all_elements.extend(make_rect("i4", 1250, 1085, 360, 115, bg="#1e1e2e", stroke="#a6e3a1",
        label="Human-in-the-Loop Governance\n• Analyst Confirm, Escalate or Override\n• Closed-Loop Retraining Feedback Loop\n• Complete Governance Audit Logging", font_size=12))

    # Arrows L5 -> L6 (Downwards)
    all_elements.extend(make_arrow("arr_5_1", 245, 1200, 245, 1260, stroke="#89b4fa"))
    all_elements.extend(make_arrow("arr_5_2", 640, 1200, 640, 1260, stroke="#89b4fa"))
    all_elements.extend(make_arrow("arr_5_3", 1035, 1200, 1035, 1260, stroke="#89b4fa"))
    all_elements.extend(make_arrow("arr_5_4", 1430, 1200, 1430, 1260, stroke="#89b4fa"))

    # ----------------------------------------------------
    # LAYER 6: DATA, STREAMING, OPERATIONS & WEBHOOK EGRESS
    # Y: 1260, H: 180 (Header Y: 1275, Cards Y: 1305)
    # ----------------------------------------------------
    all_elements.extend(make_rect("grp_l6", 40, 1260, 1600, 180, bg="#11111b", stroke="#313244", stroke_width=1.5))
    all_elements.extend(make_text_standalone("lbl_l6", 60, 1275, "LAYER 6: DATA PERSISTENCE, STREAMING EVENT BUS & OPERATIONS", font_size=13, color="#fab387"))

    all_elements.extend(make_rect("p1", 65, 1305, 360, 115, bg="#1e1e2e", stroke="#89b4fa",
        label="PostgreSQL & Redis Storage\n• Field-Level Encryption (AES-256 GCM)\n• In-Memory Feature Store & Traversal Cache\n• Continuous WAL Archiving to S3", font_size=12))
    
    all_elements.extend(make_rect("p2", 460, 1305, 360, 115, bg="#1e1e2e", stroke="#fab387",
        label="Apache Kafka Event Fabric\n• Topic: transactions.evaluated\n• Dead-Letter Queue (DLQ) Fallback\n• High-Throughput Event Fanout Bus", font_size=12))
    
    all_elements.extend(make_rect("p3", 855, 1305, 360, 115, bg="#1e1e2e", stroke="#f38ba8",
        label="High Availability & Circuit Breakers\n• Stateful Circuit Breakers (Closed/Open/Half)\n• Fallback Buffering During Dependency Failures\n• Tested Failure-Resilient Transaction Handling", font_size=12))
    
    all_elements.extend(make_rect("p4", 1250, 1305, 360, 115, bg="#1e1e2e", stroke="#a6e3a1",
        label="Observability 2.0 & SLO Monitor\n• Prometheus Metric Counters & Histograms\n• Distributed OpenTelemetry / Jaeger Tracing\n• Contractual 99.99% Availability Monitoring", font_size=12))

    # ----------------------------------------------------
    # FINAL EGRESS: SIGNED WEBHOOK NOTIFICATION (Y: 1470, H: 75)
    # ----------------------------------------------------
    all_elements.extend(make_arrow("arr_egress_1", 640, 1420, 640, 1480, stroke="#a6e3a1"))
    all_elements.extend(make_arrow("arr_egress_2", 1430, 1420, 1000, 1480, stroke="#a6e3a1"))

    all_elements.extend(make_rect(
        "egress_box", 400, 1480, 880, 75,
        bg="#181825", stroke="#a6e3a1",
        label="Signed Webhook Egress (HMAC-SHA256)\nDispatches risk.decision.created & case.created to Customer Endpoint Rails\n5x Exponential Backoff Retry Delivery Guarantee",
        font_size=13, stroke_width=2, text_color="#a6e3a1"
    ))

    data = {
        "type": "excalidraw",
        "version": 2,
        "source": "https://excalidraw.com",
        "elements": all_elements,
        "appState": {
            "viewBackgroundColor": "#11111b",
            "gridSize": None
        },
        "files": {}
    }
    
    return data

if __name__ == "__main__":
    d = build_excalidraw()
    with open("/Users/shankar/PROJECTS/Ai Risk Manager/docs/ropus-architecture.excalidraw", "w") as f:
        json.dump(d, f, indent=2)
    print("Regenerated 100% technically defensible ropus-architecture.excalidraw successfully!")
