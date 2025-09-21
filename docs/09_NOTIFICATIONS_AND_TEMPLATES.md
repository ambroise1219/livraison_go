## Notifications, Templates, Bannières (Prisma/PostgreSQL)

### Models couverts
- `Notification`, `NotificationTemplate`, `Banner`

---

### Modèles Prisma

```prisma
// Notification
model Notification {
  id       String             @id @default(uuid())
  userId   String
  type     NotificationType
  title    String?
  content  String
  status   NotificationStatus @default(PENDING)
  sentAt   DateTime?
  createdAt DateTime          @default(now())
  
  user User @relation(fields: [userId], references: [id])
  
  @@map("notifications")
}

enum NotificationType {
  SMS
  EMAIL
  PUSH
  WHATSAPP
  IN_APP
}

enum NotificationStatus {
  PENDING
  SENT
  FAILED
}

// NotificationTemplate
model NotificationTemplate {
  id        String             @id @default(uuid())
  type      NotificationType
  category  String
  language  String             @default("fr")
  title     String?
  content   String
  variables String[]           // Array de variables comme {{name}}, {{amount}}
  isActive  Boolean            @default(true)
  createdAt DateTime           @default(now())
  
  @@unique([type, category, language])
  @@map("notification_templates")
}

// Banner
model Banner {
  id             String   @id @default(uuid())
  title          String?
  description    String?
  imageUrl       String
  type           String
  targetAudience String
  platform       String
  position       String
  priority       Int      @default(0)
  isActive       Boolean  @default(true)
  startDate      DateTime @default(now())
  endDate        DateTime?
  createdAt      DateTime @default(now())
  
  @@map("banners")
}
```

---

### Services Go - Notifications

```go
type NotificationService struct {
    db           *db.PrismaClient
    smsService   SMSService
    emailService EmailService
    pushService  PushService
}

// Envoyer une notification depuis un template
func (s *NotificationService) SendNotificationFromTemplate(
    userID, category string, 
    notifType db.NotificationType, 
    language string, 
    variables map[string]string,
) error {
    // 1) Charger le template
    template, err := s.db.NotificationTemplate.FindFirst(
        db.NotificationTemplate.Category.Equals(category),
        db.NotificationTemplate.Type.Equals(notifType),
        db.NotificationTemplate.Language.Equals(language),
        db.NotificationTemplate.IsActive.Equals(true),
    ).Exec(context.Background())
    
    if err != nil {
        return err
    }
    
    // 2) Remplacer les variables dans le contenu
    content := s.replaceVariables(template.Content, variables)
    title := ""
    if template.Title != nil {
        title = s.replaceVariables(*template.Title, variables)
    }
    
    // 3) Créer la notification
    notification, err := s.db.Notification.CreateOne(
        db.Notification.UserID.Set(userID),
        db.Notification.Type.Set(notifType),
        db.Notification.Title.SetIfPresent(&title),
        db.Notification.Content.Set(content),
    ).Exec(context.Background())
    
    if err != nil {
        return err
    }
    
    // 4) Envoyer selon le type
    return s.sendNotification(notification)
}

// Remplacer les variables dans le texte
func (s *NotificationService) replaceVariables(text string, variables map[string]string) string {
    for key, value := range variables {
        text = strings.ReplaceAll(text, "{{" + key + "}}", value)
    }
    return text
}

// Envoyer la notification selon son type
func (s *NotificationService) sendNotification(notif *db.NotificationModel) error {
    var err error
    
    switch notif.Type {
    case db.NotificationTypeSMS:
        err = s.smsService.SendSMS(notif.UserID, notif.Content)
    case db.NotificationTypeEMAIL:
        err = s.emailService.SendEmail(notif.UserID, notif.Title, notif.Content)
    case db.NotificationTypePUSH:
        err = s.pushService.SendPush(notif.UserID, notif.Title, notif.Content)
    case db.NotificationTypeWHATSAPP:
        err = s.smsService.SendWhatsApp(notif.UserID, notif.Content)
    }
    
    // Mettre à jour le statut
    status := db.NotificationStatusSENT
    var sentAt *time.Time = nil
    if err != nil {
        status = db.NotificationStatusFAILED
    } else {
        now := time.Now()
        sentAt = &now
    }
    
    _, updateErr := s.db.Notification.FindUnique(
        db.Notification.ID.Equals(notif.ID),
    ).Update(
        db.Notification.Status.Set(status),
        db.Notification.SentAt.SetIfPresent(sentAt),
    ).Exec(context.Background())
    
    if updateErr != nil {
        return updateErr
    }
    
    return err
}

// Lister notifications d'un utilisateur
func (s *NotificationService) GetUserNotifications(userID string, limit int) ([]*models.Notification, error) {
    notifications, err := s.db.Notification.FindMany(
        db.Notification.UserID.Equals(userID),
    ).OrderBy(
        db.Notification.CreatedAt.Order(db.DESC),
    ).Take(limit).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaNotifications(notifications), nil
}
```

---

### Services Go - Templates

```go
type TemplateService struct {
    db *db.PrismaClient
}

// Créer un template
func (s *TemplateService) CreateTemplate(
    notifType db.NotificationType,
    category, language, title, content string,
    variables []string,
) (*models.NotificationTemplate, error) {
    template, err := s.db.NotificationTemplate.CreateOne(
        db.NotificationTemplate.Type.Set(notifType),
        db.NotificationTemplate.Category.Set(category),
        db.NotificationTemplate.Language.Set(language),
        db.NotificationTemplate.Title.SetIfPresent(&title),
        db.NotificationTemplate.Content.Set(content),
        db.NotificationTemplate.Variables.Set(variables),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaTemplateToModel(template), nil
}

// Récupérer un template
func (s *TemplateService) GetTemplate(category string, notifType db.NotificationType, language string) (*models.NotificationTemplate, error) {
    template, err := s.db.NotificationTemplate.FindFirst(
        db.NotificationTemplate.Category.Equals(category),
        db.NotificationTemplate.Type.Equals(notifType),
        db.NotificationTemplate.Language.Equals(language),
        db.NotificationTemplate.IsActive.Equals(true),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaTemplateToModel(template), nil
}
```

---

### Services Go - Bannières

```go
type BannerService struct {
    db *db.PrismaClient
}

// Créer une bannière
func (s *BannerService) CreateBanner(
    title, description, imageUrl, bannerType, targetAudience, platform, position string,
    priority int,
    startDate, endDate *time.Time,
) (*models.Banner, error) {
    banner, err := s.db.Banner.CreateOne(
        db.Banner.Title.SetIfPresent(&title),
        db.Banner.Description.SetIfPresent(&description),
        db.Banner.ImageUrl.Set(imageUrl),
        db.Banner.Type.Set(bannerType),
        db.Banner.TargetAudience.Set(targetAudience),
        db.Banner.Platform.Set(platform),
        db.Banner.Position.Set(position),
        db.Banner.Priority.Set(priority),
        db.Banner.StartDate.SetIfPresent(startDate),
        db.Banner.EndDate.SetIfPresent(endDate),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaBannerToModel(banner), nil
}

// Lister bannières actives
func (s *BannerService) GetActiveBanners(platform, targetAudience string) ([]*models.Banner, error) {
    now := time.Now()
    
    banners, err := s.db.Banner.FindMany(
        db.Banner.IsActive.Equals(true),
        db.Banner.Platform.Equals(platform),
        db.Banner.Or(
            db.Banner.TargetAudience.Equals(targetAudience),
            db.Banner.TargetAudience.Equals("ALL"),
        ),
        db.Banner.StartDate.Lte(now),
        db.Banner.Or(
            db.Banner.EndDate.IsNull(),
            db.Banner.EndDate.Gte(now),
        ),
    ).OrderBy(
        db.Banner.Priority.Order(db.DESC),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaBanners(banners), nil
}
```

---

### Middleware de permissions et index

```go
// Middleware pour notifications (utilisateur propriétaire ou admin)
func RequireNotificationAccess() gin.HandlerFunc {
    return func(c *gin.Context) {
        notifID := c.Param("notificationId")
        userID := c.GetString("user_id")
        user := c.MustGet("user").(*models.User)
        
        if user.Role == "ADMIN" {
            c.Next()
            return
        }
        
        // Vérifier la propriété
        notification, err := notificationService.GetNotification(notifID)
        if err != nil || notification.UserID != userID {
            c.JSON(403, gin.H{"error": "Access denied"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Middleware pour templates et bannières (admin/marketing)
func RequireMarketingAccess() gin.HandlerFunc {
    return RequireRole("ADMIN", "MARKETING")
}
```

**Index PostgreSQL:**
```sql
-- Notifications
CREATE INDEX notifications_user_id_idx ON notifications(user_id);
CREATE INDEX notifications_status_created_idx ON notifications(status, created_at);
CREATE INDEX notifications_user_status_idx ON notifications(user_id, status);

-- Templates
CREATE UNIQUE INDEX notification_templates_type_category_language_key 
    ON notification_templates(type, category, language);
CREATE INDEX notification_templates_active_idx ON notification_templates(is_active);

-- Bannières
CREATE INDEX banners_active_platform_audience_idx 
    ON banners(is_active, platform, target_audience);
CREATE INDEX banners_dates_idx ON banners(start_date, end_date);
CREATE INDEX banners_priority_idx ON banners(priority DESC);
```


