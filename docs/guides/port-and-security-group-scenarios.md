---
page_title: "Port & Security Group Scenarios"
subcategory: "Networking"
description: |-
  Common real-world scenarios for vnpaycloud_network_interface (port) and
  vnpaycloud_security_group / vnpaycloud_security_group_rule, with ready-to-use HCL.
---

# Port & Security Group Scenarios

Trang này tổng hợp các tình huống thường gặp khi cấu hình **port** (`vnpaycloud_network_interface`)
và **security group** (`vnpaycloud_security_group`, `vnpaycloud_security_group_rule`) trên VNPayCloud,
kèm file HCL dùng được ngay. Mỗi case bám sát schema thực tế của provider.

Tham khảo chi tiết từng resource:

- [`vnpaycloud_security_group`](../resources/security_group.md)
- [`vnpaycloud_security_group_rule`](../resources/security_group_rule.md)
- [`vnpaycloud_network_interface`](../resources/network_interface.md)
- [`vnpaycloud_network_interface_attachment`](../resources/network_interface_attachment.md)

## Ràng buộc cốt lõi cần nhớ

| Resource | Ràng buộc |
|---|---|
| `security_group` | `name` bắt buộc (unique trong project). Mỗi SG mới **tự sinh sẵn một egress allow-all** (`0.0.0.0/0`, mọi protocol) → đừng tạo lại egress trùng. `enable_log` chỉ dùng được ở zone có `can_enable_log = true`. |
| `security_group_rule` | `security_group_id`, `direction` bắt buộc. `direction` / `protocol` / `ethertype` / `port_range_min` / `port_range_max` đều **ForceNew**. **Không có `remote_group_id`** → nguồn phải khai bằng CIDR (`remote_ip_prefix`). ICMP: `port_range_min` = type, `port_range_max` = code. Bỏ `protocol` = mọi protocol; bỏ port range = mọi port. |
| `network_interface` (port) | `name` (cho phép `""`) và `subnet_id` bắt buộc; `name` / `subnet_id` / `ip_address` là **ForceNew**. Phần host của `ip_address` phải nằm trong `[16, 250]`. `reserved = true` **không destroy được** (hạ `false` + apply trước). `security_groups` nếu set thì **phải chứa System SG** và cần `port_security_enabled = true`; **không được set `[]`**. `port_security_enabled = false` xoá sạch SG → không đi kèm `security_groups`. |
| `network_interface_attachment` | `network_interface_id` và `server_id`, cả hai **ForceNew**. NIC phụ phải ở **subnet khác** với NIC chính (tránh trùng IP). |

~> **Lưu ý quan trọng:** Provider **không** hỗ trợ `remote_group_id`. Để tham chiếu "nguồn là một
security group / một tầng khác", hãy dùng **CIDR của subnet** tương ứng qua `remote_ip_prefix`
(xem [Case SG-4](#case-sg-4--mô-hình-3-tầng-web--app--db-tham-chiếu-bằng-cidr)).

## Nền tảng dùng chung

Tất cả ví dụ bên dưới giả định đã có sẵn VPC + subnet sau:

```hcl
resource "vnpaycloud_vpc" "main" {
  name = "demo-vpc"
  cidr = "10.0.0.0/16"
}

resource "vnpaycloud_subnet" "app" {
  name   = "app-subnet"
  vpc_id = vnpaycloud_vpc.main.id
  cidr   = "10.0.1.0/24"
}
```

---

## A. Security Group

### Case SG-1 — Web server cơ bản (SSH + HTTP/HTTPS)

Máy chủ web public: mở 80/443 cho mọi nơi, SSH chỉ từ dải IP văn phòng.
Không cần khai egress (SG đã có sẵn egress allow-all).

```hcl
resource "vnpaycloud_security_group" "web" {
  name        = "web-sg"
  description = "Web server - public 80 443 office-only ssh"
}

resource "vnpaycloud_security_group_rule" "web_http" {
  security_group_id = vnpaycloud_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  ethertype         = "IPv4"
  port_range_min    = 80
  port_range_max    = 80
  remote_ip_prefix  = "0.0.0.0/0"
}

resource "vnpaycloud_security_group_rule" "web_https" {
  security_group_id = vnpaycloud_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  ethertype         = "IPv4"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "0.0.0.0/0"
}

resource "vnpaycloud_security_group_rule" "web_ssh_office" {
  security_group_id = vnpaycloud_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "203.0.113.0/24" # dải IP văn phòng
  description       = "SSH from office only"
}
```

### Case SG-2 — Dải port + UDP (game/app server)

Mở một dải TCP (30000–30100) và một cổng UDP.

```hcl
resource "vnpaycloud_security_group" "app" {
  name        = "app-sg"
  description = "App with tcp port range and udp"
}

resource "vnpaycloud_security_group_rule" "app_tcp_range" {
  security_group_id = vnpaycloud_security_group.app.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = 30000
  port_range_max    = 30100
  remote_ip_prefix  = "0.0.0.0/0"
}

resource "vnpaycloud_security_group_rule" "app_udp" {
  security_group_id = vnpaycloud_security_group.app.id
  direction         = "ingress"
  protocol          = "udp"
  port_range_min    = 51820
  port_range_max    = 51820
  remote_ip_prefix  = "0.0.0.0/0"
}
```

### Case SG-3 — Cho phép ICMP (ping)

Với ICMP, `port_range_min` là **type**, `port_range_max` là **code**
(type 8 = echo-request).

```hcl
resource "vnpaycloud_security_group" "icmp" {
  name        = "icmp-sg"
  description = "Allow ping from internal"
}

resource "vnpaycloud_security_group_rule" "allow_ping" {
  security_group_id = vnpaycloud_security_group.icmp.id
  direction         = "ingress"
  protocol          = "icmp"
  port_range_min    = 8 # ICMP type: echo-request
  port_range_max    = 0 # ICMP code
  remote_ip_prefix  = "10.0.0.0/16"
}
```

### Case SG-4 — Mô hình 3 tầng web → app → db (tham chiếu bằng CIDR)

DB chỉ nhận 5432 từ tầng app; app chỉ nhận 8080 từ tầng web.
Vì provider **không có `remote_group_id`**, ta phân tầng bằng **CIDR của subnet nguồn** —
đây là điểm khác biệt quan trọng so với OpenStack/AWS.

```hcl
# Mỗi tầng một subnet để dùng CIDR làm "nguồn"
resource "vnpaycloud_subnet" "web" {
  name   = "web-subnet"
  vpc_id = vnpaycloud_vpc.main.id
  cidr   = "10.0.1.0/24"
}

resource "vnpaycloud_subnet" "app_tier" {
  name   = "app-subnet"
  vpc_id = vnpaycloud_vpc.main.id
  cidr   = "10.0.2.0/24"
}

resource "vnpaycloud_subnet" "db" {
  name   = "db-subnet"
  vpc_id = vnpaycloud_vpc.main.id
  cidr   = "10.0.3.0/24"
}

resource "vnpaycloud_security_group" "app_tier" {
  name = "tier-app-sg"
}

resource "vnpaycloud_security_group" "db_tier" {
  name = "tier-db-sg"
}

# App: nhận 8080 chỉ từ dải web
resource "vnpaycloud_security_group_rule" "app_from_web" {
  security_group_id = vnpaycloud_security_group.app_tier.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = 8080
  port_range_max    = 8080
  remote_ip_prefix  = vnpaycloud_subnet.web.cidr # 10.0.1.0/24
}

# DB: nhận 5432 chỉ từ dải app
resource "vnpaycloud_security_group_rule" "db_from_app" {
  security_group_id = vnpaycloud_security_group.db_tier.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = 5432
  port_range_max    = 5432
  remote_ip_prefix  = vnpaycloud_subnet.app_tier.cidr # 10.0.2.0/24
}
```

### Case SG-5 — Bật network logging (chỉ zone hỗ trợ)

Cần audit traffic ACCEPT. Chỉ chạy được khi `can_enable_log = true`; nếu zone không
hỗ trợ, đặt `enable_log = true` lúc create sẽ **fail và SG bị rollback** (không được tạo).

```hcl
resource "vnpaycloud_security_group" "audited" {
  name        = "audited-sg"
  description = "Security group with accept logging"
  enable_log  = true # chỉ dùng ở zone có can_enable_log = true
}
```

### Case SG-6 — Siết egress (chỉ cho ra DNS + HTTPS)

SG mới đã có sẵn một egress allow-all mà provider **không quản**. Để siết egress, bạn
xoá rule mặc định đó trên Console/API trước, rồi quản egress mong muốn bằng Terraform
như dưới (tránh để trùng `0.0.0.0/0` all-protocol với rule mặc định nếu nó còn tồn tại).

```hcl
resource "vnpaycloud_security_group" "egress_locked" {
  name        = "egress-locked-sg"
  description = "Restricted egress"
}

resource "vnpaycloud_security_group_rule" "egress_https" {
  security_group_id = vnpaycloud_security_group.egress_locked.id
  direction         = "egress"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "0.0.0.0/0"
}

resource "vnpaycloud_security_group_rule" "egress_dns" {
  security_group_id = vnpaycloud_security_group.egress_locked.id
  direction         = "egress"
  protocol          = "udp"
  port_range_min    = 53
  port_range_max    = 53
  remote_ip_prefix  = "0.0.0.0/0"
}
```

---

## B. Port (network_interface)

### Case P-1 — Port IP tĩnh

IP cố định; phần host phải trong `[16, 250]` (ví dụ `/24`: `.16`–`.250`).

```hcl
resource "vnpaycloud_network_interface" "static" {
  name        = "app-static-nic"
  subnet_id   = vnpaycloud_subnet.app.id
  ip_address  = "10.0.1.20"
  description = "Primary NIC for app server"
}
```

### Case P-2 — Port IP động

Để hệ thống tự cấp IP hợp lệ.

```hcl
resource "vnpaycloud_network_interface" "dynamic" {
  name      = "dynamic-nic"
  subnet_id = vnpaycloud_subnet.app.id
}
```

### Case P-3 — Port reserved (giữ trước IP)

Đặt trước IP để sau này gắn.

~> **Lưu ý vòng đời:** `reserved = true` **không destroy được**. Trước khi `terraform destroy`,
phải đổi `reserved = false` rồi `terraform apply`, sau đó mới xoá. Nếu không sẽ gặp lỗi
`This is a reserved port. You cannot delete.`

```hcl
resource "vnpaycloud_network_interface" "reserved_ip" {
  name       = "reserved-nic"
  subnet_id  = vnpaycloud_subnet.app.id
  ip_address = "10.0.1.30"
  reserved   = true
}
```

### Case P-4 — VIP / HA với `allowed_address_pairs` (keepalived/VRRP)

Cặp node HA cùng "lái" một Virtual IP; cần cho phép địa chỉ VIP đi qua port.

```hcl
resource "vnpaycloud_network_interface" "ha_node" {
  name       = "ha-node-nic"
  subnet_id  = vnpaycloud_subnet.app.id
  virtual_ip = true

  allowed_address_pairs {
    ip_address = "10.0.1.100" # VIP nổi giữa các node
  }
  allowed_address_pairs {
    ip_address = "10.0.2.0/24" # cho phép cả một dải (vd pod CIDR)
  }
}
```

### Case P-5 — Port gắn Security Group tùy chỉnh (bắt buộc kèm System SG)

Muốn port dùng SG riêng. **Phải** giữ System SG trong danh sách và để
`port_security_enabled = true` (mặc định). Không được set `security_groups = []`.

```hcl
data "vnpaycloud_security_group" "system" {
  name = "System Security Group"
}

resource "vnpaycloud_security_group" "web" {
  name = "web-sg"
}

resource "vnpaycloud_network_interface" "with_sg" {
  name      = "sg-nic"
  subnet_id = vnpaycloud_subnet.app.id

  security_groups = [
    data.vnpaycloud_security_group.system.id,
    vnpaycloud_security_group.web.id,
  ]
}
```

### Case P-6 — Tắt port security (NAT / virtual appliance / router ảo)

Thiết bị định tuyến cần forward gói với IP nguồn khác → tắt anti-spoof.
Khi `port_security_enabled = false`, **không** được khai báo `security_groups`.

```hcl
resource "vnpaycloud_network_interface" "appliance" {
  name                  = "nat-appliance-nic"
  subnet_id             = vnpaycloud_subnet.app.id
  port_security_enabled = false
  # KHÔNG set security_groups khi tắt port security
}
```

### Case P-7 — Port không tên (`name = ""`)

Tạo NIC ẩn danh, IP động.

```hcl
resource "vnpaycloud_network_interface" "unnamed" {
  name      = ""
  subnet_id = vnpaycloud_subnet.app.id
}
```

---

## C. Kết hợp Port + Security Group (kịch bản thực tế)

### Case C-1 — Web server: SG đặt trên Port, gắn vào Instance lúc tạo

Tạo port (đã gắn `web-sg` + System SG), rồi tạo instance dùng port đó qua
`network_interface_ids`.

```hcl
data "vnpaycloud_security_group" "system" {
  name = "System Security Group"
}

resource "vnpaycloud_security_group" "web" {
  name        = "web-sg"
  description = "Web tier"
}

resource "vnpaycloud_security_group_rule" "web_https" {
  security_group_id = vnpaycloud_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "0.0.0.0/0"
}

resource "vnpaycloud_network_interface" "web" {
  name       = "web-nic"
  subnet_id  = vnpaycloud_subnet.app.id
  ip_address = "10.0.1.50"

  security_groups = [
    data.vnpaycloud_security_group.system.id,
    vnpaycloud_security_group.web.id,
  ]
}

resource "vnpaycloud_keypair" "deployer" {
  name = "deployer-key"
}

resource "vnpaycloud_instance" "web" {
  name                  = "web-01"
  image                 = "ubuntu-22.04"
  flavor                = "s.4c8r"
  root_disk_gb          = 40
  root_disk_type        = "SSD"
  key_pair              = vnpaycloud_keypair.deployer.name
  network_interface_ids = [vnpaycloud_network_interface.web.id]
}
```

### Case C-2 — Gắn thêm NIC phụ vào instance đang chạy (khác subnet)

Instance đã có NIC chính ở subnet `app`, gắn thêm NIC ở subnet khác (vd mạng storage).
NIC phụ **bắt buộc khác subnet** với NIC chính.

```hcl
resource "vnpaycloud_subnet" "storage" {
  name   = "storage-subnet"
  vpc_id = vnpaycloud_vpc.main.id
  cidr   = "10.0.9.0/24"
}

resource "vnpaycloud_network_interface" "extra" {
  name      = "storage-nic"
  subnet_id = vnpaycloud_subnet.storage.id
}

resource "vnpaycloud_network_interface_attachment" "extra" {
  network_interface_id = vnpaycloud_network_interface.extra.id
  server_id            = vnpaycloud_instance.web.id
}
```

### Case C-3 — Cặp HA với VIP + SG + allowed_address_pairs

Hai node chạy keepalived, chia sẻ VIP `10.0.1.100`, cùng một SG cho phép cổng app nội bộ.

```hcl
data "vnpaycloud_security_group" "system" {
  name = "System Security Group"
}

resource "vnpaycloud_security_group" "ha" {
  name        = "ha-sg"
  description = "HA pair"
}

# Cho phép app port nội bộ giữa hai node
resource "vnpaycloud_security_group_rule" "ha_app" {
  security_group_id = vnpaycloud_security_group.ha.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = 6443
  port_range_max    = 6443
  remote_ip_prefix  = vnpaycloud_subnet.app.cidr
}

resource "vnpaycloud_network_interface" "node" {
  count      = 2
  name       = "ha-node-${count.index}"
  subnet_id  = vnpaycloud_subnet.app.id
  virtual_ip = true

  security_groups = [
    data.vnpaycloud_security_group.system.id,
    vnpaycloud_security_group.ha.id,
  ]

  allowed_address_pairs {
    ip_address = "10.0.1.100" # VIP dùng chung
  }
}
```

---

## Tổng hợp các "bẫy" thường gặp

1. **Không có `remote_group_id`** → phân tầng bằng CIDR ([Case SG-4](#case-sg-4--mô-hình-3-tầng-web--app--db-tham-chiếu-bằng-cidr)).
2. **Egress allow-all mặc định** sẵn có trên mọi SG mới → đừng tạo trùng.
3. **Port `security_groups` phải gồm System SG**, cần `port_security_enabled = true`, cấm `[]`.
4. **`port_security_enabled = false`** thì bỏ hẳn `security_groups`.
5. **`reserved = true`** chặn destroy → hạ `false` + apply trước.
6. **`ip_address` host ∈ [16, 250]**.
7. **NIC phụ phải khác subnet** với NIC chính.
