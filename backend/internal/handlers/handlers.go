package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"digcatalog/internal/middleware"
	"digcatalog/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	DB        *gorm.DB
	JWTSecret string
}

func New(db *gorm.DB, jwtSecret string) *Handler {
	return &Handler{DB: db, JWTSecret: jwtSecret}
}

// ---------- Auth ----------

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名和密码"})
		return
	}
	var user models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	token, err := middleware.GenerateToken(h.JWTSecret, user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func (h *Handler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":       c.MustGet("userId"),
		"username": c.MustGet("username"),
		"role":     c.MustGet("role"),
	})
}

// ---------- Sites ----------

func (h *Handler) ListSites(c *gin.Context) {
	var sites []models.Site
	if err := h.DB.Order("id desc").Find(&sites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sites)
}

func (h *Handler) GetSite(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var site models.Site
	if err := h.DB.First(&site, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工地不存在"})
		return
	}
	c.JSON(http.StatusOK, site)
}

func (h *Handler) CreateSite(c *gin.Context) {
	var site models.Site
	if err := c.ShouldBindJSON(&site); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	if site.Name == "" || site.Period == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称和时代必填"})
		return
	}
	if err := h.DB.Create(&site).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, site)
}

func (h *Handler) UpdateSite(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var site models.Site
	if err := h.DB.First(&site, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "工地不存在"})
		return
	}
	var req models.Site
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	site.Name = req.Name
	site.Period = req.Period
	site.Latitude = req.Latitude
	site.Longitude = req.Longitude
	site.Manager = req.Manager
	if err := h.DB.Save(&site).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, site)
}

func (h *Handler) DeleteSite(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var count int64
	h.DB.Model(&models.Unit{}).Where("site_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该工地下仍有探方，无法删除"})
		return
	}
	if err := h.DB.Delete(&models.Site{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// ---------- Units ----------

// validateUnitSize 校验探方平面尺寸：长宽需同时填写或同时留空，且为正数。
func validateUnitSize(u *models.Unit) string {
	if (u.LengthCm == nil) != (u.WidthCm == nil) {
		return "探方长和宽需同时填写或同时留空"
	}
	if u.LengthCm != nil && (*u.LengthCm <= 0 || *u.WidthCm <= 0) {
		return "探方长和宽需为正整数（厘米）"
	}
	return ""
}

func (h *Handler) ListUnits(c *gin.Context) {
	var units []models.Unit
	q := h.DB.Preload("Site").Order("id desc")
	if siteID := c.Query("siteId"); siteID != "" {
		q = q.Where("site_id = ?", siteID)
	}
	if err := q.Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, units)
}

func (h *Handler) GetUnit(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var unit models.Unit
	if err := h.DB.Preload("Site").First(&unit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "探方不存在"})
		return
	}
	c.JSON(http.StatusOK, unit)
}

func (h *Handler) CreateUnit(c *gin.Context) {
	var unit models.Unit
	if err := c.ShouldBindJSON(&unit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	if unit.SiteID == 0 || unit.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "所属工地和编号必填"})
		return
	}
	if msg := validateUnitSize(&unit); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	var site models.Site
	if err := h.DB.First(&site, unit.SiteID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "所属工地不存在"})
		return
	}
	if err := h.DB.Create(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.DB.Preload("Site").First(&unit, unit.ID)
	c.JSON(http.StatusCreated, unit)
}

func (h *Handler) UpdateUnit(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var unit models.Unit
	if err := h.DB.First(&unit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "探方不存在"})
		return
	}
	var req models.Unit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	if msg := validateUnitSize(&req); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	unit.SiteID = req.SiteID
	unit.Code = req.Code
	unit.DepthMin = req.DepthMin
	unit.DepthMax = req.DepthMax
	unit.LengthCm = req.LengthCm
	unit.WidthCm = req.WidthCm
	unit.StratumDesc = req.StratumDesc
	if err := h.DB.Save(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.DB.Preload("Site").First(&unit, unit.ID)
	c.JSON(http.StatusOK, unit)
}

func (h *Handler) DeleteUnit(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var count int64
	h.DB.Model(&models.Find{}).Where("unit_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该探方下仍有文物，无法删除"})
		return
	}
	if err := h.DB.Delete(&models.Unit{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// GetUnitSpotMap 返回探方平面尺寸及该探方全部带坐标文物的出土点。
// GET /units/:id/spot-map
func (h *Handler) GetUnitSpotMap(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var unit models.Unit
	if err := h.DB.Preload("Site").First(&unit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "探方不存在"})
		return
	}
	var finds []models.Find
	if err := h.DB.
		Where("unit_id = ? AND x_cm IS NOT NULL AND y_cm IS NOT NULL AND z_cm IS NOT NULL", id).
		Order("register_no asc").
		Find(&finds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type spot struct {
		ID         uint   `json:"id"`
		RegisterNo string `json:"registerNo"`
		XCm        int    `json:"xCm"`
		YCm        int    `json:"yCm"`
		ZCm        int    `json:"zCm"`
	}
	spots := make([]spot, 0, len(finds))
	for _, f := range finds {
		spots = append(spots, spot{
			ID:         f.ID,
			RegisterNo: f.RegisterNo,
			XCm:        *f.XCm,
			YCm:        *f.YCm,
			ZCm:        *f.ZCm,
		})
	}
	siteName := ""
	if unit.Site != nil {
		siteName = unit.Site.Name
	}
	c.JSON(http.StatusOK, gin.H{
		"unit": gin.H{
			"id":       unit.ID,
			"code":     unit.Code,
			"siteId":   unit.SiteID,
			"siteName": siteName,
			"lengthCm": unit.LengthCm,
			"widthCm":  unit.WidthCm,
		},
		"finds": spots,
	})
}

// ---------- Materials ----------

func (h *Handler) ListMaterials(c *gin.Context) {
	var materials []models.Material
	if err := h.DB.Order("id asc").Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, materials)
}

func (h *Handler) CreateMaterial(c *gin.Context) {
	var m models.Material
	if err := c.ShouldBindJSON(&m); err != nil || m.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称必填"})
		return
	}
	if err := h.DB.Create(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *Handler) UpdateMaterial(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var m models.Material
	if err := h.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "材质不存在"})
		return
	}
	var req models.Material
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	m.Name = req.Name
	m.Description = req.Description
	if err := h.DB.Save(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *Handler) DeleteMaterial(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.Material{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// ---------- Finds ----------

type findReq struct {
	UnitID       uint    `json:"unitId"`
	MaterialID   *uint   `json:"materialId"`
	RegisterNo   string  `json:"registerNo"`
	ArtifactType string  `json:"artifactType"`
	MaterialName string  `json:"materialName"`
	Completeness string  `json:"completeness"`
	FindDate     *string `json:"findDate"`
	Description  string  `json:"description"`
	StorageLoc   string  `json:"storageLoc"`
	XCm          *int    `json:"xCm"`
	YCm          *int    `json:"yCm"`
	ZCm          *int    `json:"zCm"`
}

// checkCoords 校验出土点坐标：三值要么全部留空（未记录），
// 要么全部填写且为非负整数。返回 hasCoord=是否带坐标, ok=是否合法。
func checkCoords(x, y, z *int) (hasCoord bool, ok bool) {
	set := 0
	for _, v := range []*int{x, y, z} {
		if v != nil {
			set++
		}
	}
	if set == 0 {
		return false, true
	}
	if set != 3 {
		return false, false
	}
	if *x < 0 || *y < 0 || *z < 0 {
		return false, false
	}
	return true, true
}

// coordConflict 检查同探方内是否已有其他文物占用该坐标组合。
func (h *Handler) coordConflict(unitID uint, excludeID uint, x, y, z int) bool {
	var count int64
	h.DB.Model(&models.Find{}).
		Where("unit_id = ? AND id <> ? AND x_cm = ? AND y_cm = ? AND z_cm = ?", unitID, excludeID, x, y, z).
		Count(&count)
	return count > 0
}

// writeFindSaveError 把保存失败翻译成前端可读的错误；坐标/登记号冲突返回 409。
func writeFindSaveError(c *gin.Context, err error) {
	if err != nil && strings.Contains(err.Error(), "Duplicate entry") {
		if strings.Contains(err.Error(), "uidx_find_unit_coord") {
			c.JSON(http.StatusConflict, gin.H{"error": "该探方内已存在相同坐标的文物"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "登记号已存在"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func parseDate(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}

func (h *Handler) ListFinds(c *gin.Context) {
	var finds []models.Find
	q := h.DB.Preload("Unit").Preload("Unit.Site").Preload("Material").Order("id desc")
	if unitID := c.Query("unitId"); unitID != "" {
		q = q.Where("unit_id = ?", unitID)
	}
	if at := c.Query("artifactType"); at != "" {
		q = q.Where("artifact_type = ?", at)
	}
	if err := q.Find(&finds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, finds)
}

func (h *Handler) GetFind(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var find models.Find
	if err := h.DB.Preload("Unit").Preload("Material").First(&find, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文物不存在"})
		return
	}
	c.JSON(http.StatusOK, find)
}

func (h *Handler) applyFindReq(find *models.Find, req *findReq) {
	find.UnitID = req.UnitID
	find.MaterialID = req.MaterialID
	find.RegisterNo = req.RegisterNo
	find.ArtifactType = req.ArtifactType
	find.MaterialName = req.MaterialName
	find.Completeness = req.Completeness
	find.FindDate = parseDate(req.FindDate)
	find.Description = req.Description
	find.StorageLoc = req.StorageLoc
	find.XCm = req.XCm
	find.YCm = req.YCm
	find.ZCm = req.ZCm
	if find.MaterialID != nil {
		var m models.Material
		if err := h.DB.First(&m, *find.MaterialID).Error; err == nil {
			find.MaterialName = m.Name
		}
	}
}

func (h *Handler) CreateFind(c *gin.Context) {
	var req findReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	if req.UnitID == 0 || req.RegisterNo == "" || req.ArtifactType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "探方、登记号、器物类型必填"})
		return
	}
	hasCoord, ok := checkCoords(req.XCm, req.YCm, req.ZCm)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "出土点坐标需三个值都填写且为非负整数（厘米）"})
		return
	}
	var unit models.Unit
	if err := h.DB.First(&unit, req.UnitID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "所属探方不存在"})
		return
	}
	if hasCoord && h.coordConflict(req.UnitID, 0, *req.XCm, *req.YCm, *req.ZCm) {
		c.JSON(http.StatusConflict, gin.H{"error": "该探方内已存在相同坐标的文物"})
		return
	}
	var find models.Find
	h.applyFindReq(&find, &req)
	if err := h.DB.Create(&find).Error; err != nil {
		writeFindSaveError(c, err)
		return
	}
	h.DB.Preload("Unit").Preload("Material").First(&find, find.ID)
	c.JSON(http.StatusCreated, find)
}

func (h *Handler) UpdateFind(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var find models.Find
	if err := h.DB.First(&find, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文物不存在"})
		return
	}
	var req findReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	hasCoord, ok := checkCoords(req.XCm, req.YCm, req.ZCm)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "出土点坐标需三个值都填写且为非负整数（厘米）"})
		return
	}
	if hasCoord && h.coordConflict(req.UnitID, find.ID, *req.XCm, *req.YCm, *req.ZCm) {
		c.JSON(http.StatusConflict, gin.H{"error": "该探方内已存在相同坐标的文物"})
		return
	}
	h.applyFindReq(&find, &req)
	if err := h.DB.Save(&find).Error; err != nil {
		writeFindSaveError(c, err)
		return
	}
	h.DB.Preload("Unit").Preload("Material").First(&find, find.ID)
	c.JSON(http.StatusOK, find)
}

// coordReq 单独更新出土点坐标；三值全空表示清除坐标。
type coordReq struct {
	XCm *int `json:"xCm"`
	YCm *int `json:"yCm"`
	ZCm *int `json:"zCm"`
}

// UpdateFindCoord 更新指定文物的出土点坐标。PUT /finds/:id/coord
func (h *Handler) UpdateFindCoord(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var find models.Find
	if err := h.DB.First(&find, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文物不存在"})
		return
	}
	var req coordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	hasCoord, ok := checkCoords(req.XCm, req.YCm, req.ZCm)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "出土点坐标需三个值都填写且为非负整数（厘米）"})
		return
	}
	if hasCoord && h.coordConflict(find.UnitID, find.ID, *req.XCm, *req.YCm, *req.ZCm) {
		c.JSON(http.StatusConflict, gin.H{"error": "该探方内已存在相同坐标的文物"})
		return
	}
	find.XCm = req.XCm
	find.YCm = req.YCm
	find.ZCm = req.ZCm
	if err := h.DB.Save(&find).Error; err != nil {
		writeFindSaveError(c, err)
		return
	}
	h.DB.Preload("Unit").Preload("Material").First(&find, find.ID)
	c.JSON(http.StatusOK, find)
}

func (h *Handler) DeleteFind(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.Find{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// ---------- Overview ----------

func (h *Handler) Overview(c *gin.Context) {
	var siteCount, unitCount, findCount int64
	h.DB.Model(&models.Site{}).Count(&siteCount)
	h.DB.Model(&models.Unit{}).Count(&unitCount)
	h.DB.Model(&models.Find{}).Count(&findCount)

	type typeStat struct {
		ArtifactType string `json:"artifactType"`
		Count        int64  `json:"count"`
	}
	var byType []typeStat
	h.DB.Model(&models.Find{}).
		Select("artifact_type as artifact_type, count(*) as count").
		Group("artifact_type").
		Scan(&byType)

	c.JSON(http.StatusOK, gin.H{
		"siteCount": siteCount,
		"unitCount": unitCount,
		"findCount": findCount,
		"byType":    byType,
	})
}
