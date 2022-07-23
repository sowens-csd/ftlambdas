# Community API
This API is used by the manager app to create people and activities in 
a community of managed users. 

## Scheduled Items
Activities, menus and potentially other things are all scheduled items.
This just means something with a date and optinoally a time associated with it, 
all day or at a particular time on a date. 

### Retrieval Patterns

1. A specific scheduled item by ID, this is possible but unusual. 
2. All scheduled items for a time range
3. All scheduled items for a tag in a time range - could get all scheduled items in a time range and then filter locally. 

