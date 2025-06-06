import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useAppDispatch } from '../../hooks/redux';
import { useWebSocket } from '../../hooks/useWebSocket';
import { 
  Calendar, 
  MapPin, 
  Plus, 
  Clock, 
  Users, 
  Settings,
  ChevronLeft,
  ChevronRight,
  MoreVertical,
  Navigation,
  X,
  Search,
  Wifi,
  WifiOff
} from 'lucide-react';
import { format, differenceInDays } from 'date-fns';
import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import toast from 'react-hot-toast';
import { reorderWaypointsThunk, removeWaypointThunk } from '../../store/thunks/trips.thunks';

interface TripPlanningPanelProps {
  isOpen: boolean;
  onClose: () => void;
}

export const TripPlanningPanel: React.FC<TripPlanningPanelProps> = ({ isOpen, onClose }) => {
  const { id: tripId } = useParams();
  const dispatch = useAppDispatch();
  
  const [activeTab, setActiveTab] = useState<'itinerary' | 'places' | 'collaborators'>('itinerary');
  const [selectedDay, setSelectedDay] = useState(0);
  const [searchQuery, setSearchQuery] = useState('');
  const [activeUsers, setActiveUsers] = useState<Array<{ id: string; name: string; avatar?: string }>>([]);
  const [typingUsers, setTypingUsers] = useState<Record<string, boolean>>({});
  
  // WebSocket integration
  const { isConnected, on } = useWebSocket({
    room: tripId ? `trip:${tripId}` : undefined,
  });
  
  // Set up real-time event listeners
  useEffect(() => {
    if (!tripId) return;
    
    const unsubscribers: Array<() => void> = [];
    
    // Listen for user join/leave events
    unsubscribers.push(
      on('user:joined', (data) => {
        if (data.tripId === tripId) {
          setActiveUsers(prev => [...prev, { id: data.userId, name: data.userName }]);
          toast(`${data.userName} joined the trip`, {
            icon: '👋',
          });
        }
      })
    );
    
    unsubscribers.push(
      on('user:left', (data) => {
        if (data.tripId === tripId) {
          setActiveUsers(prev => prev.filter(u => u.id !== data.userId));
          toast(`${data.userName} left the trip`, {
            icon: '👋',
          });
        }
      })
    );
    
    // Listen for typing events
    unsubscribers.push(
      on('user:typing', (data) => {
        if (data.context === `trip:${tripId}`) {
          setTypingUsers(prev => ({
            ...prev,
            [data.userId]: data.isTyping,
          }));
        }
      })
    );
    
    // Listen for waypoint updates
    unsubscribers.push(
      on('trip:waypoint:added', (data) => {
        if (data.tripId === tripId) {
          toast(`New place added to the trip`, {
            icon: '📍',
          });
        }
      })
    );
    
    unsubscribers.push(
      on('trip:waypoints:reordered', (data) => {
        if (data.tripId === tripId) {
          toast(`Itinerary updated`, {
            icon: '🔄',
          });
        }
      })
    );
    
    // Cleanup
    return () => {
      unsubscribers.forEach(unsub => unsub());
    };
  }, [tripId, on]);
  
  // Mock data - replace with Redux selectors
  const trip = {
    id: tripId,
    title: 'Summer Europe Adventure',
    description: 'A 2-week journey through Western Europe',
    startDate: new Date('2024-07-01'),
    endDate: new Date('2024-07-14'),
    coverImage: 'https://images.unsplash.com/photo-1499856871958-5b9627545d1a',
    status: 'planning',
    collaborators: [
      { id: '1', name: 'John Doe', role: 'owner', avatar: null },
      { id: '2', name: 'Jane Smith', role: 'editor', avatar: null },
    ],
    waypoints: [
      {
        id: 'w1',
        day: 1,
        placeId: 'p1',
        place: {
          id: 'p1',
          name: 'Eiffel Tower',
          address: 'Champ de Mars, Paris',
          category: 'attraction',
          coordinates: { lat: 48.8584, lng: 2.2945 }
        },
        arrivalTime: '10:00',
        departureTime: '12:00',
        notes: 'Book tickets in advance'
      },
      {
        id: 'w2',
        day: 1,
        placeId: 'p2',
        place: {
          id: 'p2',
          name: 'Louvre Museum',
          address: 'Rue de Rivoli, Paris',
          category: 'museum',
          coordinates: { lat: 48.8606, lng: 2.3376 }
        },
        arrivalTime: '14:00',
        departureTime: '17:00',
        notes: 'Skip the line tickets'
      }
    ]
  };

  const totalDays = differenceInDays(trip.endDate, trip.startDate) + 1;
  const daysArray = Array.from({ length: totalDays }, (_, i) => i);
  
  const waypointsForDay = trip.waypoints.filter(w => w.day === selectedDay + 1);

  const handleDragEnd = async (result: any) => {
    if (!result.destination || !tripId) return;

    const items = Array.from(waypointsForDay);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);

    // Dispatch action to reorder waypoints
    try {
      await dispatch(reorderWaypointsThunk({
        tripId,
        waypoints: items,
      })).unwrap();
    } catch (error) {
      // Error handled in thunk
    }
  };

  const handleAddPlace = () => {
    // TODO: Open place search modal
  };

  const handleRemoveWaypoint = async (waypointId: string) => {
    if (!tripId) return;
    
    try {
      await dispatch(removeWaypointThunk({
        tripId,
        waypointId,
      })).unwrap();
    } catch (error) {
      // Error handled in thunk
    }
  };

  const formatDayDate = (dayIndex: number) => {
    const date = new Date(trip.startDate);
    date.setDate(date.getDate() + dayIndex);
    return format(date, 'MMM d');
  };

  if (!isOpen) return null;

  return (
    <>
      {/* Backdrop with blur */}
      <div 
        className="fixed inset-0 top-16 bg-black/30 backdrop-blur-sm z-30"
        onClick={onClose}
      />
      
      {/* Panel Content */}
      <div className="absolute top-16 right-0 w-96 h-[calc(100vh-4rem)] bg-white shadow-2xl z-40 flex flex-col">
      {/* Header */}
      <div className="p-4 border-b bg-indigo-600 text-white">
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-xl font-bold">{trip.title}</h2>
          <div className="flex items-center gap-2">
            {/* Connection indicator */}
            <div className={`flex items-center gap-1 px-2 py-1 rounded-full text-xs ${
              isConnected ? 'bg-green-500/20' : 'bg-red-500/20'
            }`}>
              {isConnected ? (
                <>
                  <Wifi className="w-3 h-3" />
                  <span>Live</span>
                </>
              ) : (
                <>
                  <WifiOff className="w-3 h-3" />
                  <span>Offline</span>
                </>
              )}
            </div>
            <button
              onClick={onClose}
              className="p-1 hover:bg-white/20 rounded-lg transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>
        <p className="text-sm opacity-90">{trip.description}</p>
        <div className="flex items-center gap-4 mt-3 text-sm">
          <div className="flex items-center gap-1">
            <Calendar className="w-4 h-4" />
            <span>{format(trip.startDate, 'MMM d')} - {format(trip.endDate, 'MMM d, yyyy')}</span>
          </div>
          <div className="flex items-center gap-1">
            <Users className="w-4 h-4" />
            <span>{trip.collaborators.length} people</span>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex border-b">
        <button
          className={`flex-1 py-3 px-4 font-medium transition-colors ${
            activeTab === 'itinerary'
              ? 'text-indigo-600 border-b-2 border-indigo-600'
              : 'text-gray-600 hover:text-gray-900'
          }`}
          onClick={() => setActiveTab('itinerary')}
        >
          Itinerary
        </button>
        <button
          className={`flex-1 py-3 px-4 font-medium transition-colors ${
            activeTab === 'places'
              ? 'text-indigo-600 border-b-2 border-indigo-600'
              : 'text-gray-600 hover:text-gray-900'
          }`}
          onClick={() => setActiveTab('places')}
        >
          Places
        </button>
        <button
          className={`flex-1 py-3 px-4 font-medium transition-colors ${
            activeTab === 'collaborators'
              ? 'text-indigo-600 border-b-2 border-indigo-600'
              : 'text-gray-600 hover:text-gray-900'
          }`}
          onClick={() => setActiveTab('collaborators')}
        >
          People
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-hidden flex flex-col">
        {activeTab === 'itinerary' && (
          <>
            {/* Day selector */}
            <div className="p-4 border-b bg-gray-50">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-semibold text-gray-900">Trip Days</h3>
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => setSelectedDay(Math.max(0, selectedDay - 1))}
                    disabled={selectedDay === 0}
                    className="p-1 rounded hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <ChevronLeft className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => setSelectedDay(Math.min(totalDays - 1, selectedDay + 1))}
                    disabled={selectedDay === totalDays - 1}
                    className="p-1 rounded hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <ChevronRight className="w-4 h-4" />
                  </button>
                </div>
              </div>
              <div className="flex gap-2 overflow-x-auto pb-2">
                {daysArray.map((dayIndex) => (
                  <button
                    key={dayIndex}
                    onClick={() => setSelectedDay(dayIndex)}
                    className={`flex-shrink-0 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                      selectedDay === dayIndex
                        ? 'bg-indigo-600 text-white'
                        : 'bg-white text-gray-700 hover:bg-gray-100 border'
                    }`}
                  >
                    <div className="text-xs">Day {dayIndex + 1}</div>
                    <div>{formatDayDate(dayIndex)}</div>
                  </button>
                ))}
              </div>
            </div>

            {/* Waypoints for selected day */}
            <div className="flex-1 overflow-y-auto p-4">
              <DragDropContext onDragEnd={handleDragEnd}>
                <Droppable droppableId="waypoints">
                  {(provided) => (
                    <div
                      {...provided.droppableProps}
                      ref={provided.innerRef}
                      className="space-y-3"
                    >
                      {waypointsForDay.length === 0 ? (
                        <div className="text-center py-8 text-gray-500">
                          <MapPin className="w-12 h-12 mx-auto mb-3 text-gray-300" />
                          <p>No places planned for this day</p>
                          <button
                            onClick={handleAddPlace}
                            className="mt-4 text-indigo-600 hover:text-indigo-700 font-medium"
                          >
                            Add your first place
                          </button>
                        </div>
                      ) : (
                        waypointsForDay.map((waypoint, index) => (
                          <Draggable key={waypoint.id} draggableId={waypoint.id} index={index}>
                            {(provided, snapshot) => (
                              <div
                                ref={provided.innerRef}
                                {...provided.draggableProps}
                                {...provided.dragHandleProps}
                                className={`bg-white rounded-lg border p-4 ${
                                  snapshot.isDragging ? 'shadow-lg' : 'shadow-sm'
                                } hover:shadow-md transition-shadow`}
                              >
                                <div className="flex items-start justify-between">
                                  <div className="flex-1">
                                    <h4 className="font-semibold text-gray-900">
                                      {waypoint.place.name}
                                    </h4>
                                    <p className="text-sm text-gray-600 mt-1">
                                      {waypoint.place.address}
                                    </p>
                                    <div className="flex items-center gap-4 mt-2 text-sm text-gray-500">
                                      <div className="flex items-center gap-1">
                                        <Clock className="w-4 h-4" />
                                        <span>{waypoint.arrivalTime} - {waypoint.departureTime}</span>
                                      </div>
                                    </div>
                                    {waypoint.notes && (
                                      <p className="mt-2 text-sm text-gray-700 italic">
                                        {waypoint.notes}
                                      </p>
                                    )}
                                  </div>
                                  <div className="flex items-center gap-1 ml-3">
                                    <button className="p-1 hover:bg-gray-100 rounded">
                                      <Navigation className="w-4 h-4 text-gray-500" />
                                    </button>
                                    <div className="relative group">
                                      <button className="p-1 hover:bg-gray-100 rounded">
                                        <MoreVertical className="w-4 h-4 text-gray-500" />
                                      </button>
                                      <div className="absolute right-0 mt-1 w-40 bg-white rounded-lg shadow-lg border opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all">
                                        <button
                                          onClick={() => handleRemoveWaypoint(waypoint.id)}
                                          className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-gray-50"
                                        >
                                          Remove from day
                                        </button>
                                      </div>
                                    </div>
                                  </div>
                                </div>
                              </div>
                            )}
                          </Draggable>
                        ))
                      )}
                      {provided.placeholder}
                    </div>
                  )}
                </Droppable>
              </DragDropContext>

              {waypointsForDay.length > 0 && (
                <button
                  onClick={handleAddPlace}
                  className="w-full mt-4 py-3 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-indigo-400 hover:text-indigo-600 transition-colors flex items-center justify-center gap-2"
                >
                  <Plus className="w-5 h-5" />
                  Add place to this day
                </button>
              )}
            </div>
          </>
        )}

        {activeTab === 'places' && (
          <div className="flex-1 p-4">
            {/* Search bar */}
            <div className="relative mb-4">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
              <input
                type="text"
                placeholder="Search places..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
            </div>
            
            {/* Places list */}
            <div className="space-y-3">
              <p className="text-gray-500 text-sm">
                All places added to this trip will appear here
              </p>
            </div>
          </div>
        )}

        {activeTab === 'collaborators' && (
          <div className="flex-1 p-4">
            {/* Active users */}
            {activeUsers.length > 0 && (
              <div className="mb-4">
                <h4 className="text-sm font-medium text-gray-600 mb-2">Currently Active</h4>
                <div className="flex gap-2 flex-wrap">
                  {activeUsers.map((user) => (
                    <div key={user.id} className="flex items-center gap-2 px-3 py-1 bg-green-50 border border-green-200 rounded-full">
                      <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
                      <span className="text-sm text-green-800">{user.name}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
            
            {/* Collaborators list */}
            <h4 className="text-sm font-medium text-gray-600 mb-3">All Collaborators</h4>
            <div className="space-y-3">
              {trip.collaborators.map((collaborator) => {
                const isActive = activeUsers.some(u => u.id === collaborator.id);
                const isTyping = typingUsers[collaborator.id];
                
                return (
                  <div key={collaborator.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div className="flex items-center gap-3">
                      <div className="relative">
                        <div className="w-10 h-10 bg-indigo-100 rounded-full flex items-center justify-center">
                          <span className="text-indigo-600 font-semibold">
                            {collaborator.name.charAt(0).toUpperCase()}
                          </span>
                        </div>
                        {isActive && (
                          <div className="absolute bottom-0 right-0 w-3 h-3 bg-green-500 border-2 border-white rounded-full" />
                        )}
                      </div>
                      <div>
                        <p className="font-medium text-gray-900">
                          {collaborator.name}
                          {isTyping && (
                            <span className="text-xs text-gray-500 ml-2 italic">typing...</span>
                          )}
                        </p>
                        <p className="text-sm text-gray-500 capitalize">{collaborator.role}</p>
                      </div>
                    </div>
                    {collaborator.role !== 'owner' && (
                      <button className="text-gray-400 hover:text-gray-600">
                        <MoreVertical className="w-5 h-5" />
                      </button>
                    )}
                  </div>
                );
              })}
            </div>
            
            <button className="w-full mt-4 py-2 border border-indigo-600 text-indigo-600 rounded-lg hover:bg-indigo-50 transition-colors">
              Invite Collaborators
            </button>
          </div>
        )}
      </div>

      {/* Footer Actions */}
      <div className="p-4 border-t bg-gray-50">
        <button className="w-full py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors flex items-center justify-center gap-2">
          <Settings className="w-5 h-5" />
          Trip Settings
        </button>
      </div>
    </div>
    </>
  );
};