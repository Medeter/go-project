package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql" // นำเข้าไดรเวอร์ MySQL
)

// User struct สำหรับข้อมูลผู้ใช้ (รวมรหัสผ่าน)
type User struct {
	ID       int    `json:"id"`       // รหัสผู้ใช้
	Name     string `json:"name"`     // ชื่อผู้ใช้
	Password string `json:"password"` // รหัสผ่าน
	Email    string `json:"email"`    // อีเมลผู้ใช้
	Status   string `json:"status"`   // สถานะผู้ใช้
}

// Product struct สำหรับข้อมูลผลิตภัณฑ์
type Products struct {
	ID    int    `json:"g_id"`  // รหัสผลิตภัณฑ์
	Good  string `json:"good"`  // ชื่อผลิตภัณฑ์
	Price int    `json:"price"` // ราคาผลิตภัณฑ์
	Qty   int    `json:"qty"`   // จำนวนผลิตภัณฑ์
}

var db *sql.DB // ประกาศตัวแปร db เพื่อเชื่อมต่อกับฐานข้อมูล

func main() {
	var err error
	// เชื่อมต่อกับฐานข้อมูล MySQL
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/testxxx") // เปลี่ยนเป็นข้อมูลประจำตัวฐานข้อมูลของคุณ
	if err != nil {
		log.Fatal(err) // แสดงข้อผิดพลาดหากไม่สามารถเชื่อมต่อได้
	}
	defer db.Close() // ปิดการเชื่อมต่อกับฐานข้อมูลเมื่อฟังก์ชัน main เสร็จสิ้น

	// ตรวจสอบการเชื่อมต่อกับฐานข้อมูล
	if err := db.Ping(); err != nil {
		log.Fatal("ไม่สามารถเชื่อมต่อกับฐานข้อมูล: ", err) // แสดงข้อผิดพลาดหากไม่สามารถ ping ฐานข้อมูลได้
	}

	// กำหนดเส้นทางสำหรับ HTTP
	http.HandleFunc("/", redirectToLogin)                // เส้นทางหลักเพื่อ redirect ไปที่หน้า login
	http.HandleFunc("/login", handleLogin)               // เส้นทางสำหรับการจัดการคำขอเข้าสู่ระบบ
	http.HandleFunc("/products", handleProducts)         // เส้นทางสำหรับการจัดการคำขอผลิตภัณฑ์
	http.HandleFunc("/login.html", serveLoginPage)       // เส้นทางสำหรับการให้บริการหน้าเข้าสู่ระบบ
	http.HandleFunc("/products.html", serveProductsPage) // เส้นทางสำหรับการให้บริการหน้าแสดงข้อมูลผลิตภัณฑ์

	log.Println("เซิร์ฟเวอร์เริ่มต้นที่ :8080") // แสดงข้อความว่าเซิร์ฟเวอร์เริ่มทำงานที่พอร์ต 8080
	err = http.ListenAndServe(":8080", nil)     // เริ่มฟังคำขอ HTTP ที่พอร์ต 8080
	if err != nil {
		log.Fatal(err) // แสดงข้อผิดพลาดหากไม่สามารถเริ่มฟังได้
	}
}

// redirectToLogin จะทำการ redirect ไปยังหน้าเข้าสู่ระบบ
func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login.html", http.StatusFound) // เปลี่ยนเส้นทางไปยังหน้า login
}

// serveLoginPage ให้บริการหน้า HTML สำหรับเข้าสู่ระบบ
func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "login.html") // ส่งไฟล์ HTML สำหรับหน้า login
}

// serveProductsPage ให้บริการหน้า HTML สำหรับแสดงข้อมูลผลิตภัณฑ์
func serveProductsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "products.html") // ส่งไฟล์ HTML สำหรับหน้า products
}

// handleLogin จัดการคำขอเข้าสู่ระบบ
func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // ตั้งค่า Header ประเภทข้อมูลเป็น JSON

	if r.Method == http.MethodPost { // ตรวจสอบว่าเป็นคำขอ POST
		var loginUser User
		if err := json.NewDecoder(r.Body).Decode(&loginUser); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest) // ส่งกลับข้อผิดพลาด 400 Bad Request หากการ decode ล้มเหลว
			return
		}

		// ตรวจสอบชื่อผู้ใช้และรหัสผ่าน
		if err := validateUser(loginUser.Name, loginUser.Password); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized) // ส่งกลับข้อผิดพลาด 401 Unauthorized
			return
		}

		// หากการเข้าสู่ระบบสำเร็จ
		w.WriteHeader(http.StatusOK)                    // ส่งสถานะ 200 OK
		json.NewEncoder(w).Encode("เข้าสู่ระบบสำเร็จ!") // ส่งข้อความเข้าสู่ระบบสำเร็จ
	}
}

// validateUser ตรวจสอบชื่อผู้ใช้และรหัสผ่าน
func validateUser(name, password string) error {
	var dbPassword string
	// ดึงรหัสผ่านจากฐานข้อมูลตามชื่อผู้ใช้
	err := db.QueryRow("SELECT password FROM users WHERE name=?", name).Scan(&dbPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง") // หากไม่พบผู้ใช้
		}
		return err // ส่งกลับข้อผิดพลาดหากมีข้อผิดพลาดในการสอบถาม
	}

	// ตรวจสอบรหัสผ่าน (ในกรณีที่มีการเข้ารหัสรหัสผ่าน ควรเปลี่ยนการเปรียบเทียบนี้)
	if dbPassword != password {
		return fmt.Errorf("ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง") // หากรหัสผ่านไม่ถูกต้อง
	}
	return nil // หากชื่อผู้ใช้และรหัสผ่านถูกต้อง
}

// handleProducts จัดการคำขอผลิตภัณฑ์
func handleProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // ตั้งค่า Header ประเภทข้อมูลเป็น JSON

	switch r.Method {
	case http.MethodGet:
		getProducts(w, r) // แสดงข้อมูลผลิตภัณฑ์
	case http.MethodPost:
		addProduct(w, r) // เพิ่มผลิตภัณฑ์
	case http.MethodPut:
		updateProduct(w, r) // แก้ไขผลิตภัณฑ์
	case http.MethodDelete:
		deleteProduct(w, r) // ลบผลิตภัณฑ์
	default:
		http.Error(w, "วิธีการ HTTP ไม่ถูกต้อง", http.StatusMethodNotAllowed) // ส่งกลับข้อผิดพลาดหากวิธีการ HTTP ไม่ถูกต้อง
	}
}

// getProducts ดึงข้อมูลผลิตภัณฑ์ทั้งหมด
func getProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT g_id, good, price, qty FROM product") // สอบถามข้อมูลผลิตภัณฑ์ทั้งหมด
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // ส่งกลับข้อผิดพลาด 500 Internal Server Error
		return
	}
	defer rows.Close() // ปิดการเชื่อมต่อ rows เมื่อเสร็จสิ้น

	var products []Products // สร้าง slice สำหรับเก็บผลิตภัณฑ์
	for rows.Next() {
		var product Products
		// ดึงข้อมูลจากผลลัพธ์การสอบถาม
		if err := rows.Scan(&product.ID, &product.Good, &product.Price, &product.Qty); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) // ส่งกลับข้อผิดพลาด 500 Internal Server Error
			return
		}
		products = append(products, product) // เพิ่มผลิตภัณฑ์ลงใน slice
	}

	w.WriteHeader(http.StatusOK)        // ส่งกลับสถานะ 200 OK
	json.NewEncoder(w).Encode(products) // ส่งกลับข้อมูลผลิตภัณฑ์
}

// addProduct เพิ่มผลิตภัณฑ์ใหม่
func addProduct(w http.ResponseWriter, r *http.Request) {
	var product Products
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // ส่งกลับข้อผิดพลาด 400 Bad Request
		return
	}

	// เพิ่มผลิตภัณฑ์ลงในฐานข้อมูล
	_, err := db.Exec("INSERT INTO product (good, price, qty) VALUES (?, ?, ?)", product.Good, product.Price, product.Qty)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // ส่งกลับข้อผิดพลาด 500 Internal Server Error
		return
	}

	w.WriteHeader(http.StatusCreated)                  // ส่งกลับสถานะ 201 Created
	json.NewEncoder(w).Encode("เพิ่มผลิตภัณฑ์สำเร็จ!") // ส่งข้อความเพิ่มผลิตภัณฑ์สำเร็จ
}

// updateProduct แก้ไขผลิตภัณฑ์ที่มีอยู่
func updateProduct(w http.ResponseWriter, r *http.Request) {
	var product Products
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // ส่งกลับข้อผิดพลาด 400 Bad Request
		return
	}

	// อัปเดตผลิตภัณฑ์ในฐานข้อมูล
	_, err := db.Exec("UPDATE product SET good=?, price=?, qty=? WHERE g_id=?", product.Good, product.Price, product.Qty, product.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // ส่งกลับข้อผิดพลาด 500 Internal Server Error
		return
	}

	w.WriteHeader(http.StatusOK)                       // ส่งกลับสถานะ 200 OK
	json.NewEncoder(w).Encode("แก้ไขผลิตภัณฑ์สำเร็จ!") // ส่งข้อความแก้ไขผลิตภัณฑ์สำเร็จ
}

// deleteProduct ลบผลิตภัณฑ์
func deleteProduct(w http.ResponseWriter, r *http.Request) {
	var product Products
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // ส่งกลับข้อผิดพลาด 400 Bad Request
		return
	}

	// ลบผลิตภัณฑ์จากฐานข้อมูล
	_, err := db.Exec("DELETE FROM product WHERE g_id=?", product.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // ส่งกลับข้อผิดพลาด 500 Internal Server Error
		return
	}

	w.WriteHeader(http.StatusOK)                    // ส่งกลับสถานะ 200 OK
	json.NewEncoder(w).Encode("ลบผลิตภัณฑ์สำเร็จ!") // ส่งข้อความลบผลิตภัณฑ์สำเร็จ
}
