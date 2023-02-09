#ifndef OBJECT_H
#include <gst/gst.h>
#include <stdbool.h>

void gst_set_string_property(GstObject *element, char *property_name, char *property_value);

void gst_set_bool_property(GstObject *element, char *property_name, bool *property_value);

void gst_set_int_property(GstObject *element, char *property_name, int property_value);

void gst_set_caps_property(GstObject *element, char *property_name, GstCaps *property_value);

#endif // OBJECT_H