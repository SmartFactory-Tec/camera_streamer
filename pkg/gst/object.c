#include "object.h"

void gst_set_string_property(GstObject *element, char *property_name, char *property_value) {
    g_object_set(element, property_name, property_value, NULL);
}

void gst_set_bool_property(GstObject *element, char *property_name, bool *property_value) {
    g_object_set(element, property_name, *property_value ? TRUE : FALSE, NULL);
}

void gst_set_int_property(GstObject *element, char *property_name, int property_value) {
    g_object_set(element, property_name, property_value, NULL);
}

void gst_set_caps_property(GstObject *element, char *property_name, GstCaps *property_value) {
    g_object_set(element, property_name, property_value, NULL);
}