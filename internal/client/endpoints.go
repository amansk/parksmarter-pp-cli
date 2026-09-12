package client

// DefaultBaseURL is the production consumer API host (APK v4.4.0 verified Sep 2026).
const DefaultBaseURL = "https://apiv3.parksmarter.com"

// Verified ASP.NET Web API action paths (existence confirmed without auth token).
const (
	PathApplicationValidity = "/api/ApplicationValidity"
	PathStateList           = "/api/State/List"

	PathLoginPhonePassword = "/api/Auth/LoginWithPhoneAndPassword"
	PathLoginCachedToken   = "/api/Auth/LoginWithCachedAuthToken"
	PathLoginApple         = "/api/Auth/LoginWithApple"
	PathRefreshToken       = "/api/Auth/RefreshToken"
	PathLogoutAllDevices   = "/api/Auth/LogOutAllDevices"

	PathUserDetail = "/api/User/GetUserDetail"

	PathCards    = "/api/Card/GetCards"
	PathVehicles = "/api/Vehicle/GetVehicles"

	PathMetersNearby            = "/api/MeterList/GetMetersNearby"
	PathMetersByLatLng          = "/api/MeterList/GetMetersByLatLng"
	PathMetersByScannerCode     = "/api/MeterList/GetMetersByScannerCode"
	PathMetersByZoneSearch      = "/api/MeterList/GetMetersByZoneNameSearch"
	PathMetersByZoneSpaceSearch = "/api/MeterList/GetMetersByZoneOrSpaceNameSearch"
	PathMetersByLocationAddress = "/api/MeterList/GetMetersByLocationAddress"
	PathMetersLimitedByLocation = "/api/MeterList/GetLimitedMetersByLocation"
	PathMeterByNumber           = "/api/MeterList/GetMeterByNumber"

	PathParkingEstimateItems = "/api/ParkingEstimate/GetParkingEstimateItems"
	PathParkingEstimateMulti = "/api/ParkingEstimate/GetParkingEstimateMulti"

	PathSessionsActive = "/api/ParkingSession/GetActiveParkingSessions"
	PathSessionsPast   = "/api/ParkingSession/GetPastParkingSessions"

	PathStartSession  = "/api/ParkingSession/StartParkingSession"
	PathExtendSession = "/api/ParkingSession/ExtendSession"
	PathStopSession   = "/api/ParkingSession/StopSession"
)
