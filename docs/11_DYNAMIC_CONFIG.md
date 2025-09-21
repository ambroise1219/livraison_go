## Configuration dynamique (Pricing, Peak, Weather, Limits)

### Tables couvertes
- `PlatformConfig`, `PeakHoursConfig`, `WeatherConfig`, `UserLimits`, `AppVersion`

---

### Schémas (extraits)

```sql
-- PlatformConfig
DEFINE TABLE PlatformConfig SCHEMAFULL;
DEFINE FIELD configType ON TABLE PlatformConfig TYPE string ASSERT $value INSIDE [
  "COMMISSION","SERVICE_FEE","PEAK_MULTIPLIER","WEATHER_MULTIPLIER",
  "DISTANCE_MULTIPLIER","TIME_MULTIPLIER","CANCELLATION_FEE","MINIMUM_FARE"
];
DEFINE FIELD vehicleType ON TABLE PlatformConfig TYPE string ASSERT $value INSIDE ["MOTO","VOITURE","CAMIONNETTE","ALL"];
DEFINE FIELD serviceZoneId ON TABLE PlatformConfig TYPE record<ServiceZone>;
DEFINE FIELD value ON TABLE PlatformConfig TYPE float;
DEFINE FIELD percentage ON TABLE PlatformConfig TYPE float;
DEFINE FIELD isActive ON TABLE PlatformConfig TYPE bool DEFAULT true;
DEFINE FIELD validFrom ON TABLE PlatformConfig TYPE datetime DEFAULT time::now();
DEFINE FIELD validTo ON TABLE PlatformConfig TYPE option<datetime>;

-- PeakHoursConfig
DEFINE TABLE PeakHoursConfig SCHEMAFULL;
DEFINE FIELD startTime ON TABLE PeakHoursConfig TYPE string; -- "18:00"
DEFINE FIELD endTime ON TABLE PeakHoursConfig TYPE string;   -- "20:00"
DEFINE FIELD days ON TABLE PeakHoursConfig TYPE option<array<string>>;
DEFINE FIELD multiplier ON TABLE PeakHoursConfig TYPE float DEFAULT 1.0;
DEFINE FIELD serviceZoneId ON TABLE PeakHoursConfig TYPE record<ServiceZone>;
DEFINE FIELD vehicleType ON TABLE PeakHoursConfig TYPE string;
DEFINE FIELD isActive ON TABLE PeakHoursConfig TYPE bool DEFAULT true;

-- WeatherConfig
DEFINE TABLE WeatherConfig SCHEMAFULL;
DEFINE FIELD weatherType ON TABLE WeatherConfig TYPE string;
DEFINE FIELD intensity ON TABLE WeatherConfig TYPE string;
DEFINE FIELD multiplier ON TABLE WeatherConfig TYPE float DEFAULT 1.0;
DEFINE FIELD serviceZoneId ON TABLE WeatherConfig TYPE record<ServiceZone>;
DEFINE FIELD vehicleType ON TABLE WeatherConfig TYPE string;
DEFINE FIELD isActive ON TABLE WeatherConfig TYPE bool DEFAULT true;

-- UserLimits
DEFINE TABLE UserLimits SCHEMAFULL;
DEFINE FIELD userId ON TABLE UserLimits TYPE record<User>;
DEFINE FIELD limitType ON TABLE UserLimits TYPE string;
DEFINE FIELD limitValue ON TABLE UserLimits TYPE float;
DEFINE FIELD currentValue ON TABLE UserLimits TYPE float DEFAULT 0;
DEFINE FIELD isActive ON TABLE UserLimits TYPE bool DEFAULT true;

-- AppVersion
DEFINE TABLE AppVersion SCHEMAFULL;
DEFINE FIELD platform ON TABLE AppVersion TYPE string;
DEFINE FIELD version ON TABLE AppVersion TYPE string;
DEFINE FIELD isRequired ON TABLE AppVersion TYPE bool DEFAULT false;
```

---

### Requêtes & utilisation

```sql
-- Multiplier dynamique (peak + weather)
LET $m = 1.0;

LET $peak = SELECT multiplier FROM PeakHoursConfig
  WHERE isActive = true AND vehicleType IN [$vehicleType, "ALL"] AND serviceZoneId = $zone
  LIMIT 1;
IF $peak != [] THEN LET $m = $m * $peak[0].multiplier END;

LET $w = SELECT multiplier FROM WeatherConfig
  WHERE isActive = true AND vehicleType IN [$vehicleType, "ALL"] AND serviceZoneId = $zone
  LIMIT 1;
IF $w != [] THEN LET $m = $m * $w[0].multiplier END;

RETURN { multiplier: $m };
```

---

### Permissions & Index

```sql
DEFINE TABLE PlatformConfig SCHEMAFULL
  PERMISSIONS FOR select WHERE isActive = true OR $auth.role IN ["ADMIN","GESTIONNAIRE"];

DEFINE TABLE PeakHoursConfig SCHEMAFULL
  PERMISSIONS FOR select WHERE isActive = true OR $auth.role IN ["ADMIN","GESTIONNAIRE"];

DEFINE TABLE WeatherConfig SCHEMAFULL
  PERMISSIONS FOR select WHERE isActive = true OR $auth.role IN ["ADMIN","GESTIONNAIRE"];

DEFINE TABLE UserLimits SCHEMAFULL
  PERMISSIONS
    FOR select WHERE userId = $auth.id OR $auth.role IN ["ADMIN","SUPPORT"],
    FOR create, update WHERE userId = $auth.id OR $auth.role IN ["ADMIN","SUPPORT"],
    FOR delete WHERE $auth.role = "ADMIN";
```

Recommandés:
- `PlatformConfig.configType, vehicleType, serviceZoneId, validFrom, validTo`
- `PeakHoursConfig.serviceZoneId, vehicleType, days`
- `WeatherConfig.serviceZoneId, vehicleType, weatherType, intensity`
- `UserLimits.userId, limitType, isActive`


